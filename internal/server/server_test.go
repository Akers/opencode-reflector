package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/akers/opencode-reflector/internal/config"
	"github.com/akers/opencode-reflector/internal/model"
)

// mockStore implements store.Store for testing
type mockStore struct {
	sessions    []model.SessionMetrics
	logs        []model.ReflectionLog
	watermark   time.Time
	cleanupDays int
}

func (m *mockStore) InsertSession(ctx context.Context, metrics *model.SessionMetrics) error {
	m.sessions = append(m.sessions, *metrics)
	return nil
}
func (m *mockStore) GetSessionsByDateRange(ctx context.Context, start, end time.Time) ([]model.SessionMetrics, error) {
	return m.sessions, nil
}
func (m *mockStore) GetSessionsByStatus(ctx context.Context, status model.TaskStatus) ([]model.SessionMetrics, error) {
	return m.sessions, nil
}
func (m *mockStore) InsertAgentParticipations(ctx context.Context, sessionID string, participations []model.AgentParticipation) error {
	return nil
}
func (m *mockStore) InsertToolCalls(ctx context.Context, sessionID string, calls []model.ToolCallRecord) error {
	return nil
}
func (m *mockStore) InsertReflectionLog(ctx context.Context, log *model.ReflectionLog) error {
	m.logs = append(m.logs, *log)
	return nil
}
func (m *mockStore) GetWatermark(ctx context.Context) (time.Time, error) {
	return m.watermark, nil
}
func (m *mockStore) UpdateWatermark(ctx context.Context, t time.Time) error {
	m.watermark = t
	return nil
}
func (m *mockStore) Cleanup(ctx context.Context, days int) error {
	m.cleanupDays = days
	return nil
}
func (m *mockStore) Close() error { return nil }

// Helper to create a test server
func setupTestServer(t *testing.T) (*Server, *mockStore) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	mock := &mockStore{}

	srv := NewServer(cfg, mock, tmpDir, "test")
	return srv, mock
}

// TestHealthEndpoint tests GET /api/v1/health
func TestHealthEndpoint(t *testing.T) {
	srv, _ := setupTestServer(t)

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp model.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("Expected status 'ok', got %s", resp.Status)
	}
	if resp.Version != "test" {
		t.Errorf("Expected version 'test', got %s", resp.Version)
	}
}

// TestStatsEndpoint tests GET /api/v1/stats
func TestStatsEndpoint(t *testing.T) {
	srv, _ := setupTestServer(t)

	req, _ := http.NewRequest("GET", "/api/v1/stats", nil)
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp model.StatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.TotalReflections != 0 {
		t.Errorf("Expected TotalReflections 0, got %d", resp.TotalReflections)
	}
	if resp.TotalTokensConsumed != 0 {
		t.Errorf("Expected TotalTokensConsumed 0, got %d", resp.TotalTokensConsumed)
	}
	if resp.TotalSessionsAnalyzed != 0 {
		t.Errorf("Expected TotalSessionsAnalyzed 0, got %d", resp.TotalSessionsAnalyzed)
	}
}

// TestReflectEndpoint tests POST /api/v1/reflect with valid request
func TestReflectEndpoint(t *testing.T) {
	srv, mock := setupTestServer(t)

	agentName := "test-agent"
	session := model.CanonicalSession{
		ID:       "session-123",
		ToolType: model.AgentToolTypeOpencode,
		Title:    &agentName,
		Messages: []model.CanonicalMessage{
			{
				Role:      model.MessageRoleHuman,
				Content:   "Hello",
				Timestamp: time.Now().Format(time.RFC3339),
			},
			{
				Role:      model.MessageRoleAgent,
				Content:   "Hi there!",
				Timestamp: time.Now().Format(time.RFC3339),
				AgentName: &agentName,
			},
		},
		ToolCalls: []model.CanonicalToolCall{
			{
				Type:     model.ToolCallTypeTool,
				Name:     "bash",
				Success:  true,
				CalledAt: time.Now().Format(time.RFC3339),
			},
		},
	}

	reqBody := model.ReflectRequest{
		TriggerType:   "MANUAL",
		TriggerDetail: "test",
		Sessions:      []model.CanonicalSession{session},
		Config: model.ReflectConfig{
			SentimentEnabled: false,
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/reflect", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp model.ReflectResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Status != "SUCCESS" {
		t.Errorf("Expected status 'SUCCESS', got %s", resp.Status)
	}

	// Check that session was stored
	if len(mock.sessions) != 1 {
		t.Errorf("Expected 1 session stored, got %d", len(mock.sessions))
	}

	// Check that reflection log was stored
	if len(mock.logs) != 1 {
		t.Errorf("Expected 1 reflection log, got %d", len(mock.logs))
	}
}

// TestReflectWithSessions tests POST /api/v1/reflect with 2 sessions
func TestReflectWithSessions(t *testing.T) {
	srv, mock := setupTestServer(t)

	agentName := "test-agent"
	session1 := model.CanonicalSession{
		ID:       "session-1",
		ToolType: model.AgentToolTypeOpencode,
		Title:    &agentName,
		Messages: []model.CanonicalMessage{
			{
				Role:      model.MessageRoleHuman,
				Content:   "Hello",
				Timestamp: time.Now().Format(time.RFC3339),
			},
		},
	}
	session2 := model.CanonicalSession{
		ID:       "session-2",
		ToolType: model.AgentToolTypeClaudecode,
		Title:    &agentName,
		Messages: []model.CanonicalMessage{
			{
				Role:      model.MessageRoleHuman,
				Content:   "Hi",
				Timestamp: time.Now().Format(time.RFC3339),
			},
		},
	}

	reqBody := model.ReflectRequest{
		TriggerType:   "MANUAL",
		TriggerDetail: "test",
		Sessions:      []model.CanonicalSession{session1, session2},
		Config: model.ReflectConfig{
			SentimentEnabled: false,
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/reflect", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp model.ReflectResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.SessionsAnalyzed != 2 {
		t.Errorf("Expected SessionsAnalyzed 2, got %d", resp.SessionsAnalyzed)
	}

	if len(mock.sessions) != 2 {
		t.Errorf("Expected 2 sessions stored, got %d", len(mock.sessions))
	}
}

// TestReflectInvalidBody tests POST /api/v1/reflect with invalid JSON
func TestReflectInvalidBody(t *testing.T) {
	srv, _ := setupTestServer(t)

	req, _ := http.NewRequest("POST", "/api/v1/reflect", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestCleanupEndpoint tests POST /api/v1/cleanup
func TestCleanupEndpoint(t *testing.T) {
	srv, mock := setupTestServer(t)

	reqBody := model.CleanupRequest{Days: 30}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/cleanup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	if mock.cleanupDays != 30 {
		t.Errorf("Expected cleanupDays 30, got %d", mock.cleanupDays)
	}
}

// TestCleanupDefaultDays tests POST /api/v1/cleanup with days=0 uses config default
func TestCleanupDefaultDays(t *testing.T) {
	srv, mock := setupTestServer(t)

	// Config default is 90 days
	reqBody := model.CleanupRequest{Days: 0}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/cleanup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	if mock.cleanupDays != 90 {
		t.Errorf("Expected cleanupDays 90 (config default), got %d", mock.cleanupDays)
	}
}

// TestReflectWithSentiment tests that sentiment analysis is read from metadata when enabled
func TestReflectWithSentiment(t *testing.T) {
	srv, mock := setupTestServer(t)

	agentName := "test-agent"
	session := model.CanonicalSession{
		ID:       "session-sentiment",
		ToolType: model.AgentToolTypeOpencode,
		Title:    &agentName,
		Metadata: map[string]interface{}{
			"sentiment": map[string]interface{}{
				"negative_ratio": 0.2,
				"attitude_score": 7.5,
				"approval_ratio": 0.8,
			},
		},
		Messages: []model.CanonicalMessage{
			{
				Role:      model.MessageRoleHuman,
				Content:   "I love this!",
				Timestamp: time.Now().Format(time.RFC3339),
			},
			{
				Role:      model.MessageRoleAgent,
				Content:   "Great!",
				Timestamp: time.Now().Format(time.RFC3339),
				AgentName: &agentName,
			},
		},
	}

	reqBody := model.ReflectRequest{
		TriggerType:   "MANUAL",
		TriggerDetail: "sentiment-test",
		Sessions:      []model.CanonicalSession{session},
		Config: model.ReflectConfig{
			SentimentEnabled: true, // Enabled
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/reflect", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Check that session metrics have sentiment values from metadata
	if len(mock.sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(mock.sessions))
	}

	savedSession := mock.sessions[0]
	if savedSession.HumanNegativeRatio != 0.2 {
		t.Errorf("Expected HumanNegativeRatio 0.2, got %v", savedSession.HumanNegativeRatio)
	}
	if savedSession.HumanAttitudeScore != 7.5 {
		t.Errorf("Expected HumanAttitudeScore 7.5, got %v", savedSession.HumanAttitudeScore)
	}
	if savedSession.HumanApprovalRatio != 0.8 {
		t.Errorf("Expected HumanApprovalRatio 0.8, got %v", savedSession.HumanApprovalRatio)
	}
}

// TestReflectReportPath tests that report is generated with correct path
func TestReflectReportPath(t *testing.T) {
	srv, _ := setupTestServer(t)

	agentName := "test-agent"
	session := model.CanonicalSession{
		ID:       "session-report",
		ToolType: model.AgentToolTypeOpencode,
		Title:    &agentName,
		Messages: []model.CanonicalMessage{
			{
				Role:      model.MessageRoleHuman,
				Content:   "Hello",
				Timestamp: time.Now().Format(time.RFC3339),
			},
		},
	}

	reqBody := model.ReflectRequest{
		TriggerType:   "MANUAL",
		TriggerDetail: "report-test",
		Sessions:      []model.CanonicalSession{session},
		Config: model.ReflectConfig{
			SentimentEnabled: false,
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/reflect", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp model.ReflectResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Report path should contain reports directory and date
	expectedDir := filepath.Join(srv.reflectorDir, "reports")
	if !bytes.Contains([]byte(resp.ReportPath), []byte(expectedDir)) {
		t.Errorf("ReportPath should contain %s, got %s", expectedDir, resp.ReportPath)
	}
	if !bytes.Contains([]byte(resp.ReportPath), []byte("dayreport-")) {
		t.Errorf("ReportPath should contain 'dayreport-', got %s", resp.ReportPath)
	}
}