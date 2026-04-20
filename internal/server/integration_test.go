package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/akers/opencode-reflector/internal/config"
	"github.com/akers/opencode-reflector/internal/model"
	"github.com/akers/opencode-reflector/internal/server"
	"github.com/akers/opencode-reflector/internal/store"
)

// TestIntegrationReflectFlow tests the complete reflection flow:
// POST /api/v1/reflect → ExtractAll → DetermineTaskStatus → GenerateReport → SaveReport → SQLite
func TestIntegrationReflectFlow(t *testing.T) {
	// Setup temp directory
	tmpDir := t.TempDir()
	reflectorDir := filepath.Join(tmpDir, ".reflector")
	if err := os.MkdirAll(reflectorDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, sub := range []string{"data", "data/fallback", "reports", "logs", "prompts", "hooks"} {
		os.MkdirAll(filepath.Join(reflectorDir, sub), 0755)
	}

	// Setup real SQLite store
	dbPath := filepath.Join(reflectorDir, "data", "reflector.db")
	ctx := context.Background()
	s, err := store.NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	// Create server (will use metadata for L2 classification and sentiment)
	cfg := config.DefaultConfig()
	srv := server.NewServer(cfg, s, reflectorDir, "test-integration")

	// Create test HTTP server
	ts := httptest.NewServer(srv.Routes())
	defer ts.Close()

	// Build request with a session containing completed task keywords
	reqBody := model.ReflectRequest{
		TriggerType:   string(model.TriggerTypeManual),
		TriggerDetail: "MANUAL_TEST",
		Sessions: []model.CanonicalSession{
			{
				ID:       "sess_integration_001",
				ToolType: model.AgentToolTypeOpencode,
				Title:    strPtr("Integration test session"),
				Messages: []model.CanonicalMessage{
					{
						Role:      model.MessageRoleHuman,
						Content:   "帮我实现一个登录功能",
						Timestamp: time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
					},
					{
						Role:      model.MessageRoleAgent,
						AgentName: strPtr("main"),
						Content:   "好的，我来帮你实现登录功能。任务已完成！",
						Timestamp: time.Now().Format(time.RFC3339),
						Metadata: map[string]interface{}{
							"prompt_tokens":     float64(1500),
							"completion_tokens": float64(800),
						},
					},
				},
				ToolCalls: []model.CanonicalToolCall{
					{
						Type:    model.ToolCallTypeTool,
						Name:    "read",
						Success: true,
						CalledAt: time.Now().Add(-3 * time.Minute).Format(time.RFC3339),
						DurationMs: int64Ptr(120),
					},
					{
						Type:    model.ToolCallTypeMCP,
						Name:    "filesystem",
						Success: true,
						CalledAt: time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
						DurationMs: int64Ptr(200),
					},
				},
			},
		},
		Config: model.ReflectConfig{
			SentimentEnabled: false,
		},
	}

	body, _ := json.Marshal(reqBody)

	// Send request
	resp, err := http.Post(ts.URL+"/api/v1/reflect", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(respBody))
	}

	var reflectResp model.ReflectResponse
	if err := json.NewDecoder(resp.Body).Decode(&reflectResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response
	if reflectResp.Status != "SUCCESS" {
		t.Errorf("Expected SUCCESS, got %s", reflectResp.Status)
	}
	if reflectResp.SessionsAnalyzed != 1 {
		t.Errorf("Expected 1 session analyzed, got %d", reflectResp.SessionsAnalyzed)
	}
	if reflectResp.ReportPath == "" {
		t.Error("Expected non-empty report path")
	}

	// Verify SQLite data
	sessions, err := s.GetSessionsByDateRange(ctx, time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to query sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session in DB, got %d", len(sessions))
	}

	saved := sessions[0]
	if saved.SessionID != "sess_integration_001" {
		t.Errorf("Expected session ID sess_integration_001, got %s", saved.SessionID)
	}
	if saved.TaskStatus != string(model.TaskStatusCompleted) {
		t.Errorf("Expected COMPLETED status (L3 keyword match), got %s", saved.TaskStatus)
	}
	if saved.ToolType != "opencode" {
		t.Errorf("Expected opencode tool type, got %s", saved.ToolType)
	}
	if saved.TotalMessageCount != 2 {
		t.Errorf("Expected 2 messages, got %v", saved.TotalMessageCount)
	}
	if saved.HumanMessageCount != 1 {
		t.Errorf("Expected 1 human message, got %v", saved.HumanMessageCount)
	}
	if saved.AgentMessageCount != 1 {
		t.Errorf("Expected 1 agent message, got %v", saved.AgentMessageCount)
	}
	if saved.ToolCallCount != 1 {
		t.Errorf("Expected 1 tool call, got %v", saved.ToolCallCount)
	}
	if saved.MCPCallCount != 1 {
		t.Errorf("Expected 1 MCP call, got %v", saved.MCPCallCount)
	}

	// Verify report file exists
	reportPath := reflectResp.ReportPath
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		t.Errorf("Report file not found: %s", reportPath)
	}

	// Verify report content
	reportContent, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("Failed to read report: %v", err)
	}
	if !bytes.Contains(reportContent, []byte("日报")) {
		t.Error("Report should contain '日报'")
	}
	if !bytes.Contains(reportContent, []byte("sess_int")) {
		t.Error("Report should contain session ID (truncated)")
	}
	if !bytes.Contains(reportContent, []byte("Token 消耗统计")) {
		t.Error("Report should contain Token statistics section")
	}
	if !bytes.Contains(reportContent, []byte("工具使用统计")) {
		t.Error("Report should contain tool usage section")
	}

	t.Logf("Report generated at: %s (%d bytes)", reportPath, len(reportContent))
	t.Logf("Session status: %s", saved.TaskStatus)
	t.Logf("Duration: %.1fs", saved.DurationSeconds)
}

// TestIntegrationHealthAndStats tests health and stats endpoints
func TestIntegrationHealthAndStats(t *testing.T) {
	tmpDir := t.TempDir()
	reflectorDir := filepath.Join(tmpDir, ".reflector")
	for _, sub := range []string{"data", "reports", "logs", "prompts", "hooks"} {
		os.MkdirAll(filepath.Join(reflectorDir, sub), 0755)
	}

	dbPath := filepath.Join(reflectorDir, "data", "reflector.db")
	ctx := context.Background()
	s, err := store.NewStore(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	cfg := config.DefaultConfig()
	srv := server.NewServer(cfg, s, reflectorDir, "test-health")
	ts := httptest.NewServer(srv.Routes())
	defer ts.Close()

	// Test health
	resp, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	var health model.HealthResponse
	json.NewDecoder(resp.Body).Decode(&health)
	if health.Status != "ok" {
		t.Errorf("Expected ok, got %s", health.Status)
	}
	if health.Version != "test-health" {
		t.Errorf("Expected test-health, got %s", health.Version)
	}

	// Test stats (initial)
	resp, err = http.Get(ts.URL + "/api/v1/stats")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var stats model.StatsResponse
	json.NewDecoder(resp.Body).Decode(&stats)
	if stats.TotalReflections != 0 {
		t.Errorf("Expected 0 reflections initially, got %d", stats.TotalReflections)
	}
}

// TestIntegrationCleanup tests cleanup endpoint with real SQLite
func TestIntegrationCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	reflectorDir := filepath.Join(tmpDir, ".reflector")
	for _, sub := range []string{"data", "reports", "logs", "prompts", "hooks"} {
		os.MkdirAll(filepath.Join(reflectorDir, sub), 0755)
	}

	dbPath := filepath.Join(reflectorDir, "data", "reflector.db")
	ctx := context.Background()
	s, err := store.NewStore(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Insert some data
	metrics := &model.SessionMetrics{
		SessionID:    "sess_old",
		ToolType:     "opencode",
		TaskStatus:   "COMPLETED",
		AnalyzedAt:   time.Now().Add(-100 * 24 * time.Hour).Format(time.RFC3339),
		StartedAt:    time.Now().Add(-100 * 24 * time.Hour).Format(time.RFC3339),
		EndedAt:      time.Now().Add(-100 * 24 * time.Hour).Format(time.RFC3339),
	}
	if err := s.InsertSession(ctx, metrics); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	srv := server.NewServer(cfg, s, reflectorDir, "test-cleanup")
	ts := httptest.NewServer(srv.Routes())
	defer ts.Close()

	// Cleanup with 30 days
	cleanupBody, _ := json.Marshal(model.CleanupRequest{Days: 30})
	resp, err := http.Post(ts.URL+"/api/v1/cleanup", "application/json", bytes.NewReader(cleanupBody))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
}

func strPtr(s string) *string  { return &s }
func int64Ptr(i int64) *int64 { return &i }
