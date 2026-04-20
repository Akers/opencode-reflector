package store

import (
	"context"
	"testing"
	"time"

	"github.com/akers/opencode-reflector/internal/model"
)

func TestNewStoreCreatesTables(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"

	store, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	// Verify all 5 tables exist
	tables := []string{
		"sessions",
		"agent_participations",
		"tool_calls",
		"reflection_logs",
		"watermarks",
	}

	for _, table := range tables {
		query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?`
		var name string
		err := store.db.QueryRowContext(ctx, query, table).Scan(&name)
		if err != nil {
			t.Errorf("Table %s does not exist: %v", table, err)
		}
	}
}

func TestNewStoreIdempotent(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"

	// First call
	store1, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("First NewStore failed: %v", err)
	}
	store1.Close()

	// Second call should not fail
	store2, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("Second NewStore failed: %v", err)
	}
	store2.Close()
}

func TestInsertSession(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"

	store, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	title := "Test Session"
	metrics := &model.SessionMetrics{
		SessionID:              "session-001",
		ToolType:              "opencode",
		Title:                 &title,
		StartedAt:             "2024-01-15T10:00:00Z",
		EndedAt:               "2024-01-15T10:30:00Z",
		DurationSeconds:       1800,
		ActiveTimeSeconds:     1500,
		HumanThinkTimeSeconds: 300,
		AgentThinkTimeSeconds: 1200,
		TotalPromptTokens:     1000,
		TotalCompletionTokens: 500,
		TotalTokens:           1500,
		ModelRequestCount:     10,
		ToolCallCount:         25,
		ToolSuccessCount:       24,
		MCPCallCount:          5,
		SkillCallCount:        3,
		ToolAvgDurationMs:     150,
		MCPAvgDurationMs:      200,
		AgentParticipationCount: 2,
		AgentNames:            "agent1,agent2",
		TotalMessageCount:     50,
		HumanMessageCount:     10,
		AgentMessageCount:     40,
		HumanParticipationRatio: 0.2,
		HumanAvgIntervalSeconds: 300,
		HumanMessageAvgChars:   150,
		HumanNegativeRatio:    0.05,
		HumanAttitudeScore:    0.85,
		HumanApprovalRatio:    0.9,
		TaskStatus:            "COMPLETED",
		AnalyzedAt:            "2024-01-15T10:35:00Z",
	}

	err = store.InsertSession(ctx, metrics)
	if err != nil {
		t.Fatalf("InsertSession failed: %v", err)
	}

	// Verify by querying back
	sessions, err := store.GetSessionsByStatus(ctx, model.TaskStatusCompleted)
	if err != nil {
		t.Fatalf("GetSessionsByStatus failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	s := sessions[0]
	if s.SessionID != "session-001" {
		t.Errorf("Expected SessionID 'session-001', got '%s'", s.SessionID)
	}
	if s.ToolType != "opencode" {
		t.Errorf("Expected ToolType 'opencode', got '%s'", s.ToolType)
	}
	if s.Title == nil || *s.Title != "Test Session" {
		t.Errorf("Expected Title 'Test Session', got %v", s.Title)
	}
	if s.DurationSeconds != 1800 {
		t.Errorf("Expected DurationSeconds 1800, got %f", s.DurationSeconds)
	}
	if s.ActiveTimeSeconds != 1500 {
		t.Errorf("Expected ActiveTimeSeconds 1500, got %f", s.ActiveTimeSeconds)
	}
	if s.HumanThinkTimeSeconds != 300 {
		t.Errorf("Expected HumanThinkTimeSeconds 300, got %f", s.HumanThinkTimeSeconds)
	}
	if s.AgentThinkTimeSeconds != 1200 {
		t.Errorf("Expected AgentThinkTimeSeconds 1200, got %f", s.AgentThinkTimeSeconds)
	}
	if s.TotalPromptTokens != 1000 {
		t.Errorf("Expected TotalPromptTokens 1000, got %f", s.TotalPromptTokens)
	}
	if s.TotalCompletionTokens != 500 {
		t.Errorf("Expected TotalCompletionTokens 500, got %f", s.TotalCompletionTokens)
	}
	if s.TotalTokens != 1500 {
		t.Errorf("Expected TotalTokens 1500, got %f", s.TotalTokens)
	}
	if s.ModelRequestCount != 10 {
		t.Errorf("Expected ModelRequestCount 10, got %f", s.ModelRequestCount)
	}
	if s.TotalMessageCount != 50 {
		t.Errorf("Expected TotalMessageCount 50, got %f", s.TotalMessageCount)
	}
	if s.HumanMessageCount != 10 {
		t.Errorf("Expected HumanMessageCount 10, got %f", s.HumanMessageCount)
	}
	if s.AgentMessageCount != 40 {
		t.Errorf("Expected AgentMessageCount 40, got %f", s.AgentMessageCount)
	}
	if s.HumanParticipationRatio != 0.2 {
		t.Errorf("Expected HumanParticipationRatio 0.2, got %f", s.HumanParticipationRatio)
	}
	if s.HumanAvgIntervalSeconds != 300 {
		t.Errorf("Expected HumanAvgIntervalSeconds 300, got %f", s.HumanAvgIntervalSeconds)
	}
	if s.HumanMessageAvgChars != 150 {
		t.Errorf("Expected HumanMessageAvgChars 150, got %f", s.HumanMessageAvgChars)
	}
	if s.ToolCallCount != 25 {
		t.Errorf("Expected ToolCallCount 25, got %f", s.ToolCallCount)
	}
	if s.ToolSuccessCount != 24 {
		t.Errorf("Expected ToolSuccessCount 24, got %f", s.ToolSuccessCount)
	}
	if s.MCPCallCount != 5 {
		t.Errorf("Expected MCPCallCount 5, got %f", s.MCPCallCount)
	}
	if s.SkillCallCount != 3 {
		t.Errorf("Expected SkillCallCount 3, got %f", s.SkillCallCount)
	}
	if s.ToolAvgDurationMs != 150 {
		t.Errorf("Expected ToolAvgDurationMs 150, got %f", s.ToolAvgDurationMs)
	}
	if s.MCPAvgDurationMs != 200 {
		t.Errorf("Expected MCPAvgDurationMs 200, got %f", s.MCPAvgDurationMs)
	}
	if s.HumanNegativeRatio != 0.05 {
		t.Errorf("Expected HumanNegativeRatio 0.05, got %f", s.HumanNegativeRatio)
	}
	if s.HumanAttitudeScore != 0.85 {
		t.Errorf("Expected HumanAttitudeScore 0.85, got %f", s.HumanAttitudeScore)
	}
	if s.HumanApprovalRatio != 0.9 {
		t.Errorf("Expected HumanApprovalRatio 0.9, got %f", s.HumanApprovalRatio)
	}
	if s.AgentNames != "agent1,agent2" {
		t.Errorf("Expected AgentNames 'agent1,agent2', got '%s'", s.AgentNames)
	}
	if s.AgentParticipationCount != 2 {
		t.Errorf("Expected AgentParticipationCount 2, got %f", s.AgentParticipationCount)
	}
}

func TestInsertSessionWithNA(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"

	store, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	metrics := &model.SessionMetrics{
		SessionID:              "session-002",
		ToolType:              "opencode",
		StartedAt:             "2024-01-16T10:00:00Z",
		EndedAt:               "2024-01-16T10:30:00Z",
		TotalPromptTokens:     -1, // N/A
		TotalCompletionTokens: -1, // N/A
		TotalTokens:           -1, // N/A
		ModelRequestCount:     -1, // N/A
		ToolAvgDurationMs:     -1, // N/A
		MCPAvgDurationMs:      -1, // N/A
		HumanAvgIntervalSeconds: -1, // N/A
		HumanAttitudeScore:    -1, // N/A
		HumanApprovalRatio:    -1, // N/A
		TotalMessageCount:     20,
		HumanMessageCount:     5,
		AgentMessageCount:     15,
		TaskStatus:            "COMPLETED",
	}

	err = store.InsertSession(ctx, metrics)
	if err != nil {
		t.Fatalf("InsertSession failed: %v", err)
	}

	// Verify by querying back
	sessions, err := store.GetSessionsByStatus(ctx, model.TaskStatusCompleted)
	if err != nil {
		t.Fatalf("GetSessionsByStatus failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	s := sessions[0]
	if s.TotalPromptTokens != 0 {
		t.Errorf("Expected TotalPromptTokens 0 (NULL), got %f", s.TotalPromptTokens)
	}
	if s.TotalCompletionTokens != 0 {
		t.Errorf("Expected TotalCompletionTokens 0 (NULL), got %f", s.TotalCompletionTokens)
	}
	if s.TotalTokens != 0 {
		t.Errorf("Expected TotalTokens 0 (NULL), got %f", s.TotalTokens)
	}
	if s.ModelRequestCount != 0 {
		t.Errorf("Expected ModelRequestCount 0 (NULL), got %f", s.ModelRequestCount)
	}
	if s.ToolAvgDurationMs != 0 {
		t.Errorf("Expected ToolAvgDurationMs 0 (NULL), got %f", s.ToolAvgDurationMs)
	}
	if s.MCPAvgDurationMs != 0 {
		t.Errorf("Expected MCPAvgDurationMs 0 (NULL), got %f", s.MCPAvgDurationMs)
	}
	if s.HumanAvgIntervalSeconds != 0 {
		t.Errorf("Expected HumanAvgIntervalSeconds 0 (NULL), got %f", s.HumanAvgIntervalSeconds)
	}
	if s.HumanAttitudeScore != 0 {
		t.Errorf("Expected HumanAttitudeScore 0 (NULL), got %f", s.HumanAttitudeScore)
	}
	if s.HumanApprovalRatio != 0 {
		t.Errorf("Expected HumanApprovalRatio 0 (NULL), got %f", s.HumanApprovalRatio)
	}
}

func TestGetSessionsByDateRange(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"

	store, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	// Insert 3 sessions at different dates
	sessions := []model.SessionMetrics{
		{
			SessionID:      "session-010",
			ToolType:      "opencode",
			StartedAt:     "2024-01-10T10:00:00Z",
			EndedAt:       "2024-01-10T10:30:00Z",
			TaskStatus:    "COMPLETED",
		},
		{
			SessionID:      "session-015",
			ToolType:      "opencode",
			StartedAt:     "2024-01-15T10:00:00Z",
			EndedAt:       "2024-01-15T10:30:00Z",
			TaskStatus:    "COMPLETED",
		},
		{
			SessionID:      "session-020",
			ToolType:      "opencode",
			StartedAt:     "2024-01-20T10:00:00Z",
			EndedAt:       "2024-01-20T10:30:00Z",
			TaskStatus:    "COMPLETED",
		},
	}

	for _, s := range sessions {
		err := store.InsertSession(ctx, &s)
		if err != nil {
			t.Fatalf("InsertSession failed: %v", err)
		}
	}

	// Query middle range (Jan 12 to Jan 17)
	start := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 17, 23, 59, 59, 0, time.UTC)

	result, err := store.GetSessionsByDateRange(ctx, start, end)
	if err != nil {
		t.Fatalf("GetSessionsByDateRange failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 session in range, got %d", len(result))
	}

	if result[0].SessionID != "session-015" {
		t.Errorf("Expected session-015, got %s", result[0].SessionID)
	}
}

func TestGetSessionsByStatus(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"

	store, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	// Insert sessions with different statuses
	sessions := []model.SessionMetrics{
		{
			SessionID:   "session-completed",
			ToolType:    "opencode",
			StartedAt:   "2024-01-15T10:00:00Z",
			EndedAt:     "2024-01-15T10:30:00Z",
			TaskStatus:  "COMPLETED",
		},
		{
			SessionID:   "session-interrupted",
			ToolType:    "opencode",
			StartedAt:   "2024-01-15T11:00:00Z",
			EndedAt:     "2024-01-15T11:30:00Z",
			TaskStatus:  "INTERRUPTED",
		},
		{
			SessionID:   "session-completed-2",
			ToolType:    "opencode",
			StartedAt:   "2024-01-15T12:00:00Z",
			EndedAt:     "2024-01-15T12:30:00Z",
			TaskStatus:  "COMPLETED",
		},
	}

	for _, s := range sessions {
		err := store.InsertSession(ctx, &s)
		if err != nil {
			t.Fatalf("InsertSession failed: %v", err)
		}
	}

	// Query only COMPLETED
	result, err := store.GetSessionsByStatus(ctx, model.TaskStatusCompleted)
	if err != nil {
		t.Fatalf("GetSessionsByStatus failed: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 COMPLETED sessions, got %d", len(result))
	}
}

func TestInsertAgentParticipations(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"

	store, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	// Insert a session first
	session := &model.SessionMetrics{
		SessionID:   "session-agents",
		ToolType:    "opencode",
		StartedAt:   "2024-01-15T10:00:00Z",
		EndedAt:     "2024-01-15T10:30:00Z",
		TaskStatus:  "COMPLETED",
	}
	err = store.InsertSession(ctx, session)
	if err != nil {
		t.Fatalf("InsertSession failed: %v", err)
	}

	// Insert agent participations
	participations := []model.AgentParticipation{
		{SessionID: "session-agents", AgentName: "agent-1", MessageCount: 20, ToolCallCount: 10},
		{SessionID: "session-agents", AgentName: "agent-2", MessageCount: 15, ToolCallCount: 8},
	}

	err = store.InsertAgentParticipations(ctx, "session-agents", participations)
	if err != nil {
		t.Fatalf("InsertAgentParticipations failed: %v", err)
	}

	// Verify by querying
	var count int
	err = store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM agent_participations WHERE session_id = ?", "session-agents").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 participations, got %d", count)
	}
}

func TestInsertToolCalls(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"

	store, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	// Insert a session first
	session := &model.SessionMetrics{
		SessionID:   "session-tools",
		ToolType:    "opencode",
		StartedAt:   "2024-01-15T10:00:00Z",
		EndedAt:     "2024-01-15T10:30:00Z",
		TaskStatus:  "COMPLETED",
	}
	err = store.InsertSession(ctx, session)
	if err != nil {
		t.Fatalf("InsertSession failed: %v", err)
	}

	// Insert tool calls of 3 types
	calls := []model.ToolCallRecord{
		{SessionID: "session-tools", Type: "TOOL", Name: "read_file", DurationMs: 100, Success: true, CalledAt: "2024-01-15T10:05:00Z"},
		{SessionID: "session-tools", Type: "MCP", Name: "mcp_server", DurationMs: 200, Success: true, CalledAt: "2024-01-15T10:10:00Z"},
		{SessionID: "session-tools", Type: "SKILL", Name: "skill_invoke", DurationMs: 150, Success: false, CalledAt: "2024-01-15T10:15:00Z"},
	}

	err = store.InsertToolCalls(ctx, "session-tools", calls)
	if err != nil {
		t.Fatalf("InsertToolCalls failed: %v", err)
	}

	// Verify by querying
	var toolCount, mcpCount, skillCount int
	store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tool_calls WHERE session_id = ? AND call_type = 'TOOL'", "session-tools").Scan(&toolCount)
	store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tool_calls WHERE session_id = ? AND call_type = 'MCP'", "session-tools").Scan(&mcpCount)
	store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tool_calls WHERE session_id = ? AND call_type = 'SKILL'", "session-tools").Scan(&skillCount)

	if toolCount != 1 {
		t.Errorf("Expected 1 TOOL call, got %d", toolCount)
	}
	if mcpCount != 1 {
		t.Errorf("Expected 1 MCP call, got %d", mcpCount)
	}
	if skillCount != 1 {
		t.Errorf("Expected 1 SKILL call, got %d", skillCount)
	}
}

func TestInsertReflectionLog(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"

	store, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	log := &model.ReflectionLog{
		TriggerType:      "TIME",
		TriggerDetail:   "scheduled",
		SessionsAnalyzed: 10,
		ReflectorTokens:  5000,
		DurationMs:       30000,
		ReportPath:       "/reports/2024-01-15.json",
		Status:           "SUCCESS",
		TriggeredAt:      "2024-01-15T10:00:00Z",
	}

	err = store.InsertReflectionLog(ctx, log)
	if err != nil {
		t.Fatalf("InsertReflectionLog failed: %v", err)
	}

	// Verify by querying
	var count int
	err = store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reflection_logs WHERE trigger_type = 'TIME'").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 reflection log, got %d", count)
	}
}

func TestWatermarkCRUD(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"

	store, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	// Initial state: watermark should be zero time
	wm, err := store.GetWatermark(ctx)
	if err != nil {
		t.Fatalf("GetWatermark failed: %v", err)
	}
	if !wm.IsZero() {
		t.Errorf("Expected zero time for initial watermark, got %v", wm)
	}

	// Update watermark
	newTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	err = store.UpdateWatermark(ctx, newTime)
	if err != nil {
		t.Fatalf("UpdateWatermark failed: %v", err)
	}

	// Get watermark
	wm, err = store.GetWatermark(ctx)
	if err != nil {
		t.Fatalf("GetWatermark failed: %v", err)
	}
	if wm.Unix() != newTime.Unix() {
		t.Errorf("Expected %v, got %v", newTime, wm)
	}

	// Update again
	newTime2 := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	err = store.UpdateWatermark(ctx, newTime2)
	if err != nil {
		t.Fatalf("UpdateWatermark failed: %v", err)
	}

	wm, err = store.GetWatermark(ctx)
	if err != nil {
		t.Fatalf("GetWatermark failed: %v", err)
	}
	if wm.Unix() != newTime2.Unix() {
		t.Errorf("Expected %v, got %v", newTime2, wm)
	}
}

func TestCleanup(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"

	store, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	// Insert sessions: 30 days ago, 1 day ago, today
	now := time.Now()
	sessions := []model.SessionMetrics{
		{
			SessionID:   "session-old",
			ToolType:    "opencode",
			StartedAt:   now.AddDate(0, 0, -30).Format(time.RFC3339),
			EndedAt:     now.AddDate(0, 0, -30).Add(30 * time.Minute).Format(time.RFC3339),
			TaskStatus:  "COMPLETED",
		},
		{
			SessionID:   "session-recent",
			ToolType:    "opencode",
			StartedAt:   now.AddDate(0, 0, -1).Format(time.RFC3339),
			EndedAt:     now.AddDate(0, 0, -1).Add(30 * time.Minute).Format(time.RFC3339),
			TaskStatus:  "COMPLETED",
		},
		{
			SessionID:   "session-today",
			ToolType:    "opencode",
			StartedAt:   now.Format(time.RFC3339),
			EndedAt:     now.Add(30 * time.Minute).Format(time.RFC3339),
			TaskStatus:  "COMPLETED",
		},
	}

	for _, s := range sessions {
		err := store.InsertSession(ctx, &s)
		if err != nil {
			t.Fatalf("InsertSession failed: %v", err)
		}
	}

	// Cleanup sessions older than 7 days
	err = store.Cleanup(ctx, 7)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify: only recent and today should remain
	result, err := store.GetSessionsByDateRange(ctx, now.AddDate(0, 0, -7), now.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("GetSessionsByDateRange failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 sessions after cleanup, got %d", len(result))
	}
}

func TestCleanupCascade(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"

	store, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	// Insert a session
	now := time.Now()
	session := &model.SessionMetrics{
		SessionID:   "session-cascade",
		ToolType:    "opencode",
		StartedAt:   now.AddDate(0, 0, -30).Format(time.RFC3339),
		EndedAt:     now.AddDate(0, 0, -30).Add(30 * time.Minute).Format(time.RFC3339),
		TaskStatus:  "COMPLETED",
	}
	err = store.InsertSession(ctx, session)
	if err != nil {
		t.Fatalf("InsertSession failed: %v", err)
	}

	// Insert agent participations
	participations := []model.AgentParticipation{
		{SessionID: "session-cascade", AgentName: "agent-1", MessageCount: 20, ToolCallCount: 10},
	}
	err = store.InsertAgentParticipations(ctx, "session-cascade", participations)
	if err != nil {
		t.Fatalf("InsertAgentParticipations failed: %v", err)
	}

	// Insert tool calls
	calls := []model.ToolCallRecord{
		{SessionID: "session-cascade", Type: "TOOL", Name: "read_file", DurationMs: 100, Success: true, CalledAt: now.Format(time.RFC3339)},
	}
	err = store.InsertToolCalls(ctx, "session-cascade", calls)
	if err != nil {
		t.Fatalf("InsertToolCalls failed: %v", err)
	}

	// Verify all exist before cleanup
	var sessionCount, partCount, toolCount int
	store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sessions WHERE id = 'session-cascade'").Scan(&sessionCount)
	store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM agent_participations WHERE session_id = 'session-cascade'").Scan(&partCount)
	store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tool_calls WHERE session_id = 'session-cascade'").Scan(&toolCount)

	if sessionCount != 1 || partCount != 1 || toolCount != 1 {
		t.Fatalf("Expected 1 each before cleanup, got session=%d, part=%d, tool=%d", sessionCount, partCount, toolCount)
	}

	// Cleanup
	err = store.Cleanup(ctx, 7)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify all are cascade deleted
	store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sessions WHERE id = 'session-cascade'").Scan(&sessionCount)
	store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM agent_participations WHERE session_id = 'session-cascade'").Scan(&partCount)
	store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tool_calls WHERE session_id = 'session-cascade'").Scan(&toolCount)

	if sessionCount != 0 || partCount != 0 || toolCount != 0 {
		t.Errorf("Expected 0 after cascade delete, got session=%d, part=%d, tool=%d", sessionCount, partCount, toolCount)
	}
}

func TestStoreClose(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"

	store, err := NewStore(ctx, dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	// Close should not panic
	err = store.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// Calling Close again should be handled gracefully
	// (sql.DB.Close is idempotent for basic usage)
}
