package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/akers/opencode-reflector/internal/config"
	"github.com/akers/opencode-reflector/internal/engine"
	"github.com/akers/opencode-reflector/internal/model"
	"github.com/akers/opencode-reflector/internal/store"
)

// Server is the core HTTP API server for opencode-reflector
type Server struct {
	config       *config.Config
	store        store.Store
	triggerMgr   *engine.TriggerManager
	hooks        map[engine.HookPoint][]string
	reflectorDir string // .reflector/ path
	version      string

	// Stats tracking
	totalReflections      int64
	totalTokensConsumed   int64
	totalSessionsAnalyzed int64
}

// NewServer creates a new Server instance
func NewServer(cfg *config.Config, s store.Store, reflectorDir string, version string) *Server {
	debounceMin := 5 * time.Minute
	tm := engine.NewTriggerManager(debounceMin)

	hooks := make(map[engine.HookPoint][]string)
	hooksDir := reflectorDir + "/hooks"
	if scanned, err := engine.ScanHooks(hooksDir); err == nil {
		hooks = scanned
	}

	return &Server{
		config:       cfg,
		store:        s,
		triggerMgr:   tm,
		hooks:        hooks,
		reflectorDir: reflectorDir,
		version:      version,
	}
}

// Routes returns a chi.Router with all API routes configured
func (s *Server) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/api/v1/reflect", s.handleReflect)
	r.Get("/api/v1/health", s.handleHealth)
	r.Get("/api/v1/stats", s.handleStats)
	r.Post("/api/v1/cleanup", s.handleCleanup)

	return r
}

// handleReflect handles POST /api/v1/reflect
func (s *Server) handleReflect(w http.ResponseWriter, r *http.Request) {
	var req model.ReflectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	triggerType := model.TriggerType(req.TriggerType)
	startTime := time.Now()

	// Execute reflection via TriggerManager
	var resp model.ReflectResponse
	err := s.triggerMgr.Trigger(triggerType, req.TriggerDetail, func(tt model.TriggerType, detail string) error {
		result, err := s.executeReflection(r.Context(), triggerType, detail, req.Sessions, req.Config)
		if err != nil {
			return err
		}
		resp = result
		return nil
	})

	if err != nil {
		log.Printf("Reflection failed: %v", err)
		http.Error(w, `{"error":"reflection failed"}`, http.StatusInternalServerError)
		return
	}

	resp.DurationMs = time.Since(startTime).Milliseconds()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// executeReflection performs the actual reflection analysis
func (s *Server) executeReflection(ctx context.Context, triggerType model.TriggerType, detail string, sessions []model.CanonicalSession, cfg model.ReflectConfig) (model.ReflectResponse, error) {
	// Execute before-reflect hook
	s.executeHooks(ctx, engine.HookBeforeReflect, nil)

	var allMetrics []model.SessionMetrics
	allParticipations := make(map[string][]model.AgentParticipation)
	allToolCalls := make(map[string][]model.ToolCallRecord)
	var totalReflectorTokens int64

	for _, session := range sessions {
		// Extract metrics
		metrics := engine.ExtractAll(&session)

		// Classify task status (from metadata)
		taskStatus := engine.DetermineTaskStatus(session)
		metrics.TaskStatus = string(taskStatus)

		// Execute after-classify hook
		s.executeHooks(ctx, engine.HookAfterClassify, metrics)

		// Sentiment analysis (from metadata, injected by Adapter)
		if cfg.SentimentEnabled {
			if sr := model.GetMetadataMap(session.Metadata, "sentiment"); sr != nil {
				metrics.HumanNegativeRatio = model.GetMetadataFloat(sr, "negative_ratio", -1)
				metrics.HumanAttitudeScore = model.GetMetadataFloat(sr, "attitude_score", -1)
				metrics.HumanApprovalRatio = model.GetMetadataFloat(sr, "approval_ratio", -1)
			}
		}

		// Save metrics to store
		if err := s.store.InsertSession(ctx, metrics); err != nil {
			log.Printf("Failed to save session metrics: %v", err)
		}

		// Extract and save agent participations
		_, participations, _ := engine.ExtractAgentMetrics(&session)
		if len(participations) > 0 {
			allParticipations[session.ID] = participations
			if err := s.store.InsertAgentParticipations(ctx, session.ID, participations); err != nil {
				log.Printf("Failed to save agent participations: %v", err)
			}
		}

		// Extract and save tool calls
		_, _, _, _, _, _, toolRecords := engine.ExtractToolMetrics(&session)
		if len(toolRecords) > 0 {
			allToolCalls[session.ID] = toolRecords
			if err := s.store.InsertToolCalls(ctx, session.ID, toolRecords); err != nil {
				log.Printf("Failed to save tool calls: %v", err)
			}
		}

		allMetrics = append(allMetrics, *metrics)
	}

	// Generate report
	reportDate := engine.GetReportDate(triggerType, time.Now())
	reportsDir := s.reflectorDir + "/reports"
	reportPath := engine.ReportFilePath(reportsDir, reportDate)

	// Execute before-report hook
	s.executeHooks(ctx, engine.HookBeforeReport, allMetrics)

	reportContent := engine.GenerateReport(reportDate, allMetrics, allParticipations, allToolCalls, totalReflectorTokens)
	if err := engine.SaveReport(reportPath, reportContent); err != nil {
		log.Printf("Failed to save report: %v", err)
	}

	// Execute after-report hook
	s.executeHooks(ctx, engine.HookAfterReport, reportPath)

	// Save reflection log
	reflectionLog := &model.ReflectionLog{
		TriggerType:      string(triggerType),
		TriggerDetail:    detail,
		SessionsAnalyzed: len(sessions),
		ReflectorTokens:  totalReflectorTokens,
		ReportPath:       reportPath,
		Status:           "SUCCESS",
		TriggeredAt:      time.Now().Format(time.RFC3339),
	}
	if err := s.store.InsertReflectionLog(ctx, reflectionLog); err != nil {
		log.Printf("Failed to save reflection log: %v", err)
	}

	// Update watermark
	if err := s.store.UpdateWatermark(ctx, time.Now()); err != nil {
		log.Printf("Failed to update watermark: %v", err)
	}

	// Update stats
	s.totalReflections++
	s.totalSessionsAnalyzed += int64(len(sessions))
	s.totalTokensConsumed += totalReflectorTokens

	// Execute after-reflect hook
	s.executeHooks(ctx, engine.HookAfterReflect, reflectionLog)

	return model.ReflectResponse{
		Status:           "SUCCESS",
		SessionsAnalyzed: len(sessions),
		ReportPath:       reportPath,
		TokensConsumed:   totalReflectorTokens,
	}, nil
}

// handleHealth handles GET /api/v1/health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := model.HealthResponse{
		Status:  "ok",
		Version: s.version,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleStats handles GET /api/v1/stats
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	resp := model.StatsResponse{
		TotalReflections:      s.totalReflections,
		TotalTokensConsumed:   s.totalTokensConsumed,
		TotalSessionsAnalyzed: s.totalSessionsAnalyzed,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleCleanup handles POST /api/v1/cleanup
func (s *Server) handleCleanup(w http.ResponseWriter, r *http.Request) {
	var req model.CleanupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Days <= 0 {
		req.Days = s.config.Retention.Days
	}

	if err := s.store.Cleanup(r.Context(), req.Days); err != nil {
		log.Printf("Cleanup failed: %v", err)
		http.Error(w, `{"error":"cleanup failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "days": strconv.Itoa(req.Days)})
}

// executeHooks executes all hooks for a given hook point
func (s *Server) executeHooks(ctx context.Context, hookPoint engine.HookPoint, input interface{}) {
	scripts, ok := s.hooks[hookPoint]
	if !ok {
		return
	}
	for _, script := range scripts {
		if err := engine.ExecuteHook(ctx, script, input, nil); err != nil {
			log.Printf("Hook %s (%s) failed: %v", hookPoint, script, err)
		}
	}
}