package model

import (
	"encoding/json"
	"testing"
)

func int64Ptr(v int64) *int64 { return &v }

func TestSessionMetricsNAValues(t *testing.T) {
	// Verify that -1 represents N/A for all float64 fields
	m := SessionMetrics{
		SessionID: "test-session",

		// Time metrics (M-001~004)
		DurationSeconds:       -1,
		ActiveTimeSeconds:     -1,
		HumanThinkTimeSeconds: -1,
		AgentThinkTimeSeconds: -1,

		// Token metrics (M-005~008)
		TotalPromptTokens:     -1,
		TotalCompletionTokens: -1,
		TotalTokens:           -1,
		ModelRequestCount:     -1,

		// Tool metrics (M-009~014)
		ToolCallCount:     -1,
		ToolSuccessCount:  -1,
		MCPCallCount:      -1,
		SkillCallCount:    -1,
		ToolAvgDurationMs: -1,
		MCPAvgDurationMs:  -1,

		// Agent metrics (M-015~016)
		AgentParticipationCount: -1,

		// Message metrics (M-017~019)
		TotalMessageCount: -1,
		HumanMessageCount: -1,
		AgentMessageCount: -1,

		// Engagement metrics (M-020~022)
		HumanParticipationRatio:  -1,
		HumanAvgIntervalSeconds: -1,
		HumanMessageAvgChars:     -1,

		// Sentiment metrics (M-023~025)
		HumanNegativeRatio: -1,
		HumanAttitudeScore: -1,
		HumanApprovalRatio: -1,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Failed to marshal SessionMetrics: %v", err)
	}

	var unmarshaled SessionMetrics
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal SessionMetrics: %v", err)
	}

	// Verify all float64 fields are -1
	if unmarshaled.DurationSeconds != -1 {
		t.Errorf("DurationSeconds: expected -1, got %v", unmarshaled.DurationSeconds)
	}
	if unmarshaled.ActiveTimeSeconds != -1 {
		t.Errorf("ActiveTimeSeconds: expected -1, got %v", unmarshaled.ActiveTimeSeconds)
	}
	if unmarshaled.HumanThinkTimeSeconds != -1 {
		t.Errorf("HumanThinkTimeSeconds: expected -1, got %v", unmarshaled.HumanThinkTimeSeconds)
	}
	if unmarshaled.AgentThinkTimeSeconds != -1 {
		t.Errorf("AgentThinkTimeSeconds: expected -1, got %v", unmarshaled.AgentThinkTimeSeconds)
	}
	if unmarshaled.TotalPromptTokens != -1 {
		t.Errorf("TotalPromptTokens: expected -1, got %v", unmarshaled.TotalPromptTokens)
	}
	if unmarshaled.TotalCompletionTokens != -1 {
		t.Errorf("TotalCompletionTokens: expected -1, got %v", unmarshaled.TotalCompletionTokens)
	}
	if unmarshaled.TotalTokens != -1 {
		t.Errorf("TotalTokens: expected -1, got %v", unmarshaled.TotalTokens)
	}
	if unmarshaled.ModelRequestCount != -1 {
		t.Errorf("ModelRequestCount: expected -1, got %v", unmarshaled.ModelRequestCount)
	}
	if unmarshaled.ToolCallCount != -1 {
		t.Errorf("ToolCallCount: expected -1, got %v", unmarshaled.ToolCallCount)
	}
	if unmarshaled.ToolSuccessCount != -1 {
		t.Errorf("ToolSuccessCount: expected -1, got %v", unmarshaled.ToolSuccessCount)
	}
	if unmarshaled.MCPCallCount != -1 {
		t.Errorf("MCPCallCount: expected -1, got %v", unmarshaled.MCPCallCount)
	}
	if unmarshaled.SkillCallCount != -1 {
		t.Errorf("SkillCallCount: expected -1, got %v", unmarshaled.SkillCallCount)
	}
	if unmarshaled.ToolAvgDurationMs != -1 {
		t.Errorf("ToolAvgDurationMs: expected -1, got %v", unmarshaled.ToolAvgDurationMs)
	}
	if unmarshaled.MCPAvgDurationMs != -1 {
		t.Errorf("MCPAvgDurationMs: expected -1, got %v", unmarshaled.MCPAvgDurationMs)
	}
	if unmarshaled.AgentParticipationCount != -1 {
		t.Errorf("AgentParticipationCount: expected -1, got %v", unmarshaled.AgentParticipationCount)
	}
	if unmarshaled.TotalMessageCount != -1 {
		t.Errorf("TotalMessageCount: expected -1, got %v", unmarshaled.TotalMessageCount)
	}
	if unmarshaled.HumanMessageCount != -1 {
		t.Errorf("HumanMessageCount: expected -1, got %v", unmarshaled.HumanMessageCount)
	}
	if unmarshaled.AgentMessageCount != -1 {
		t.Errorf("AgentMessageCount: expected -1, got %v", unmarshaled.AgentMessageCount)
	}
	if unmarshaled.HumanParticipationRatio != -1 {
		t.Errorf("HumanParticipationRatio: expected -1, got %v", unmarshaled.HumanParticipationRatio)
	}
	if unmarshaled.HumanAvgIntervalSeconds != -1 {
		t.Errorf("HumanAvgIntervalSeconds: expected -1, got %v", unmarshaled.HumanAvgIntervalSeconds)
	}
	if unmarshaled.HumanMessageAvgChars != -1 {
		t.Errorf("HumanMessageAvgChars: expected -1, got %v", unmarshaled.HumanMessageAvgChars)
	}
	if unmarshaled.HumanNegativeRatio != -1 {
		t.Errorf("HumanNegativeRatio: expected -1, got %v", unmarshaled.HumanNegativeRatio)
	}
	if unmarshaled.HumanAttitudeScore != -1 {
		t.Errorf("HumanAttitudeScore: expected -1, got %v", unmarshaled.HumanAttitudeScore)
	}
	if unmarshaled.HumanApprovalRatio != -1 {
		t.Errorf("HumanApprovalRatio: expected -1, got %v", unmarshaled.HumanApprovalRatio)
	}

	t.Logf("SessionMetrics JSON: %s", string(data))
}

func TestSessionMetricsJSONRoundTrip(t *testing.T) {
	title := "Test Session"
	m := SessionMetrics{
		SessionID: "session-123",
		ToolType:  "opencode",
		Title:     &title,

		StartedAt: "2024-01-01T10:00:00Z",
		EndedAt:   "2024-01-01T11:00:00Z",

		// Time metrics (M-001~004)
		DurationSeconds:       3600,
		ActiveTimeSeconds:     3000,
		HumanThinkTimeSeconds: 600,
		AgentThinkTimeSeconds: 2400,

		// Token metrics (M-005~008)
		TotalPromptTokens:     10000,
		TotalCompletionTokens: 5000,
		TotalTokens:           15000,
		ModelRequestCount:     50,

		// Tool metrics (M-009~014)
		ToolCallCount:     25,
		ToolSuccessCount:  23,
		MCPCallCount:      5,
		SkillCallCount:    3,
		ToolAvgDurationMs: 150,
		MCPAvgDurationMs:  200,

		// Agent metrics (M-015~016)
		AgentParticipationCount: 2,
		AgentNames:              "gpt-4,claude-3",

		// Message metrics (M-017~019)
		TotalMessageCount: 100,
		HumanMessageCount: 40,
		AgentMessageCount: 60,

		// Engagement metrics (M-020~022)
		HumanParticipationRatio:  0.4,
		HumanAvgIntervalSeconds: 90,
		HumanMessageAvgChars:     200,

		// Sentiment metrics (M-023~025)
		HumanNegativeRatio: 0.1,
		HumanAttitudeScore: 0.85,
		HumanApprovalRatio: 0.9,

		TaskStatus: "COMPLETED",
		AnalyzedAt: "2024-01-01T12:00:00Z",
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Failed to marshal SessionMetrics: %v", err)
	}

	var unmarshaled SessionMetrics
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal SessionMetrics: %v", err)
	}

	// Verify all fields
	if unmarshaled.SessionID != m.SessionID {
		t.Errorf("SessionID: expected %s, got %s", m.SessionID, unmarshaled.SessionID)
	}
	if unmarshaled.ToolType != m.ToolType {
		t.Errorf("ToolType: expected %s, got %s", m.ToolType, unmarshaled.ToolType)
	}
	if unmarshaled.Title == nil || *unmarshaled.Title != *m.Title {
		t.Errorf("Title: expected %s, got %v", *m.Title, unmarshaled.Title)
	}
	if unmarshaled.StartedAt != m.StartedAt {
		t.Errorf("StartedAt: expected %s, got %s", m.StartedAt, unmarshaled.StartedAt)
	}
	if unmarshaled.EndedAt != m.EndedAt {
		t.Errorf("EndedAt: expected %s, got %s", m.EndedAt, unmarshaled.EndedAt)
	}
	if unmarshaled.DurationSeconds != m.DurationSeconds {
		t.Errorf("DurationSeconds: expected %v, got %v", m.DurationSeconds, unmarshaled.DurationSeconds)
	}
	if unmarshaled.ActiveTimeSeconds != m.ActiveTimeSeconds {
		t.Errorf("ActiveTimeSeconds: expected %v, got %v", m.ActiveTimeSeconds, unmarshaled.ActiveTimeSeconds)
	}
	if unmarshaled.HumanThinkTimeSeconds != m.HumanThinkTimeSeconds {
		t.Errorf("HumanThinkTimeSeconds: expected %v, got %v", m.HumanThinkTimeSeconds, unmarshaled.HumanThinkTimeSeconds)
	}
	if unmarshaled.AgentThinkTimeSeconds != m.AgentThinkTimeSeconds {
		t.Errorf("AgentThinkTimeSeconds: expected %v, got %v", m.AgentThinkTimeSeconds, unmarshaled.AgentThinkTimeSeconds)
	}
	if unmarshaled.TotalPromptTokens != m.TotalPromptTokens {
		t.Errorf("TotalPromptTokens: expected %v, got %v", m.TotalPromptTokens, unmarshaled.TotalPromptTokens)
	}
	if unmarshaled.TotalCompletionTokens != m.TotalCompletionTokens {
		t.Errorf("TotalCompletionTokens: expected %v, got %v", m.TotalCompletionTokens, unmarshaled.TotalCompletionTokens)
	}
	if unmarshaled.TotalTokens != m.TotalTokens {
		t.Errorf("TotalTokens: expected %v, got %v", m.TotalTokens, unmarshaled.TotalTokens)
	}
	if unmarshaled.ModelRequestCount != m.ModelRequestCount {
		t.Errorf("ModelRequestCount: expected %v, got %v", m.ModelRequestCount, unmarshaled.ModelRequestCount)
	}
	if unmarshaled.ToolCallCount != m.ToolCallCount {
		t.Errorf("ToolCallCount: expected %v, got %v", m.ToolCallCount, unmarshaled.ToolCallCount)
	}
	if unmarshaled.ToolSuccessCount != m.ToolSuccessCount {
		t.Errorf("ToolSuccessCount: expected %v, got %v", m.ToolSuccessCount, unmarshaled.ToolSuccessCount)
	}
	if unmarshaled.MCPCallCount != m.MCPCallCount {
		t.Errorf("MCPCallCount: expected %v, got %v", m.MCPCallCount, unmarshaled.MCPCallCount)
	}
	if unmarshaled.SkillCallCount != m.SkillCallCount {
		t.Errorf("SkillCallCount: expected %v, got %v", m.SkillCallCount, unmarshaled.SkillCallCount)
	}
	if unmarshaled.ToolAvgDurationMs != m.ToolAvgDurationMs {
		t.Errorf("ToolAvgDurationMs: expected %v, got %v", m.ToolAvgDurationMs, unmarshaled.ToolAvgDurationMs)
	}
	if unmarshaled.MCPAvgDurationMs != m.MCPAvgDurationMs {
		t.Errorf("MCPAvgDurationMs: expected %v, got %v", m.MCPAvgDurationMs, unmarshaled.MCPAvgDurationMs)
	}
	if unmarshaled.AgentParticipationCount != m.AgentParticipationCount {
		t.Errorf("AgentParticipationCount: expected %v, got %v", m.AgentParticipationCount, unmarshaled.AgentParticipationCount)
	}
	if unmarshaled.AgentNames != m.AgentNames {
		t.Errorf("AgentNames: expected %s, got %s", m.AgentNames, unmarshaled.AgentNames)
	}
	if unmarshaled.TotalMessageCount != m.TotalMessageCount {
		t.Errorf("TotalMessageCount: expected %v, got %v", m.TotalMessageCount, unmarshaled.TotalMessageCount)
	}
	if unmarshaled.HumanMessageCount != m.HumanMessageCount {
		t.Errorf("HumanMessageCount: expected %v, got %v", m.HumanMessageCount, unmarshaled.HumanMessageCount)
	}
	if unmarshaled.AgentMessageCount != m.AgentMessageCount {
		t.Errorf("AgentMessageCount: expected %v, got %v", m.AgentMessageCount, unmarshaled.AgentMessageCount)
	}
	if unmarshaled.HumanParticipationRatio != m.HumanParticipationRatio {
		t.Errorf("HumanParticipationRatio: expected %v, got %v", m.HumanParticipationRatio, unmarshaled.HumanParticipationRatio)
	}
	if unmarshaled.HumanAvgIntervalSeconds != m.HumanAvgIntervalSeconds {
		t.Errorf("HumanAvgIntervalSeconds: expected %v, got %v", m.HumanAvgIntervalSeconds, unmarshaled.HumanAvgIntervalSeconds)
	}
	if unmarshaled.HumanMessageAvgChars != m.HumanMessageAvgChars {
		t.Errorf("HumanMessageAvgChars: expected %v, got %v", m.HumanMessageAvgChars, unmarshaled.HumanMessageAvgChars)
	}
	if unmarshaled.HumanNegativeRatio != m.HumanNegativeRatio {
		t.Errorf("HumanNegativeRatio: expected %v, got %v", m.HumanNegativeRatio, unmarshaled.HumanNegativeRatio)
	}
	if unmarshaled.HumanAttitudeScore != m.HumanAttitudeScore {
		t.Errorf("HumanAttitudeScore: expected %v, got %v", m.HumanAttitudeScore, unmarshaled.HumanAttitudeScore)
	}
	if unmarshaled.HumanApprovalRatio != m.HumanApprovalRatio {
		t.Errorf("HumanApprovalRatio: expected %v, got %v", m.HumanApprovalRatio, unmarshaled.HumanApprovalRatio)
	}
	if unmarshaled.TaskStatus != m.TaskStatus {
		t.Errorf("TaskStatus: expected %s, got %s", m.TaskStatus, unmarshaled.TaskStatus)
	}
	if unmarshaled.AnalyzedAt != m.AnalyzedAt {
		t.Errorf("AnalyzedAt: expected %s, got %s", m.AnalyzedAt, unmarshaled.AnalyzedAt)
	}

	t.Logf("SessionMetrics JSON: %s", string(data))
}

func TestTaskStatusConstants(t *testing.T) {
	if TaskStatusCompleted != "COMPLETED" {
		t.Errorf("TaskStatusCompleted: expected COMPLETED, got %s", TaskStatusCompleted)
	}
	if TaskStatusInterrupted != "INTERRUPTED" {
		t.Errorf("TaskStatusInterrupted: expected INTERRUPTED, got %s", TaskStatusInterrupted)
	}
	if TaskStatusUncertain != "UNCERTAIN" {
		t.Errorf("TaskStatusUncertain: expected UNCERTAIN, got %s", TaskStatusUncertain)
	}
	if TaskStatusAbandoned != "ABANDONED" {
		t.Errorf("TaskStatusAbandoned: expected ABANDONED, got %s", TaskStatusAbandoned)
	}
}

func TestTriggerTypeConstants(t *testing.T) {
	if TriggerTypeTime != "TIME" {
		t.Errorf("TriggerTypeTime: expected TIME, got %s", TriggerTypeTime)
	}
	if TriggerTypeManual != "MANUAL" {
		t.Errorf("TriggerTypeManual: expected MANUAL, got %s", TriggerTypeManual)
	}
	if TriggerTypeEvents != "EVENTS" {
		t.Errorf("TriggerTypeEvents: expected EVENTS, got %s", TriggerTypeEvents)
	}
}

func TestReflectRequestJSONRoundTrip(t *testing.T) {
	title := "Test Session"
	session := CanonicalSession{
		ID:       "session-123",
		ToolType: AgentToolTypeOpencode,
		Title:    &title,
		Messages: []CanonicalMessage{
			{
				Role:      MessageRoleHuman,
				Content:   "Hello, world!",
				Timestamp: "2024-01-01T10:00:00Z",
			},
		},
		ToolCalls: []CanonicalToolCall{
			{
				Type:       ToolCallTypeTool,
				Name:       "bash",
				DurationMs: int64Ptr(100),
				Success:    true,
				CalledAt:   "2024-01-01T10:00:01Z",
			},
		},
	}

	req := ReflectRequest{
		TriggerType:   "MANUAL",
		TriggerDetail: "user-triggered",
		Sessions:      []CanonicalSession{session},
		Config: ReflectConfig{
			SentimentEnabled: true,
			SentimentSource:  "llm",
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal ReflectRequest: %v", err)
	}

	var unmarshaled ReflectRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ReflectRequest: %v", err)
	}

	if unmarshaled.TriggerType != req.TriggerType {
		t.Errorf("TriggerType: expected %s, got %s", req.TriggerType, unmarshaled.TriggerType)
	}
	if unmarshaled.TriggerDetail != req.TriggerDetail {
		t.Errorf("TriggerDetail: expected %s, got %s", req.TriggerDetail, unmarshaled.TriggerDetail)
	}
	if len(unmarshaled.Sessions) != 1 {
		t.Errorf("Sessions: expected 1, got %d", len(unmarshaled.Sessions))
	}
	if unmarshaled.Sessions[0].ID != session.ID {
		t.Errorf("Session.ID: expected %s, got %s", session.ID, unmarshaled.Sessions[0].ID)
	}
	if unmarshaled.Config.SentimentSource != req.Config.SentimentSource {
		t.Errorf("Config.SentimentSource: expected %s, got %s", req.Config.SentimentSource, unmarshaled.Config.SentimentSource)
	}
	if unmarshaled.Config.SentimentEnabled != req.Config.SentimentEnabled {
		t.Errorf("Config.SentimentEnabled: expected %v, got %v", req.Config.SentimentEnabled, unmarshaled.Config.SentimentEnabled)
	}

	t.Logf("ReflectRequest JSON: %s", string(data))
}

func TestReflectResponseJSON(t *testing.T) {
	resp := ReflectResponse{
		Status:           "SUCCESS",
		SessionsAnalyzed: 10,
		ReportPath:       "/reports/2024-01-01 reflection",
		TokensConsumed:   50000,
		DurationMs:       1234,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal ReflectResponse: %v", err)
	}

	var unmarshaled ReflectResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ReflectResponse: %v", err)
	}

	if unmarshaled.Status != resp.Status {
		t.Errorf("Status: expected %s, got %s", resp.Status, unmarshaled.Status)
	}
	if unmarshaled.SessionsAnalyzed != resp.SessionsAnalyzed {
		t.Errorf("SessionsAnalyzed: expected %d, got %d", resp.SessionsAnalyzed, unmarshaled.SessionsAnalyzed)
	}
	if unmarshaled.ReportPath != resp.ReportPath {
		t.Errorf("ReportPath: expected %s, got %s", resp.ReportPath, unmarshaled.ReportPath)
	}
	if unmarshaled.TokensConsumed != resp.TokensConsumed {
		t.Errorf("TokensConsumed: expected %d, got %d", resp.TokensConsumed, unmarshaled.TokensConsumed)
	}
	if unmarshaled.DurationMs != resp.DurationMs {
		t.Errorf("DurationMs: expected %d, got %d", resp.DurationMs, unmarshaled.DurationMs)
	}

	t.Logf("ReflectResponse JSON: %s", string(data))
}

func TestAgentParticipationJSON(t *testing.T) {
	ap := AgentParticipation{
		SessionID:     "session-123",
		AgentName:     "gpt-4",
		MessageCount:  50,
		ToolCallCount: 25,
	}

	data, err := json.Marshal(ap)
	if err != nil {
		t.Fatalf("Failed to marshal AgentParticipation: %v", err)
	}

	var unmarshaled AgentParticipation
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal AgentParticipation: %v", err)
	}

	if unmarshaled.SessionID != ap.SessionID {
		t.Errorf("SessionID: expected %s, got %s", ap.SessionID, unmarshaled.SessionID)
	}
	if unmarshaled.AgentName != ap.AgentName {
		t.Errorf("AgentName: expected %s, got %s", ap.AgentName, unmarshaled.AgentName)
	}
	if unmarshaled.MessageCount != ap.MessageCount {
		t.Errorf("MessageCount: expected %d, got %d", ap.MessageCount, unmarshaled.MessageCount)
	}
	if unmarshaled.ToolCallCount != ap.ToolCallCount {
		t.Errorf("ToolCallCount: expected %d, got %d", ap.ToolCallCount, unmarshaled.ToolCallCount)
	}

	t.Logf("AgentParticipation JSON: %s", string(data))
}

func TestToolCallRecordJSON(t *testing.T) {
	tcr := ToolCallRecord{
		SessionID:  "session-123",
		Type:       "TOOL",
		Name:       "bash",
		DurationMs: 150,
		Success:    true,
		CalledAt:   "2024-01-01T10:00:00Z",
	}

	data, err := json.Marshal(tcr)
	if err != nil {
		t.Fatalf("Failed to marshal ToolCallRecord: %v", err)
	}

	var unmarshaled ToolCallRecord
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ToolCallRecord: %v", err)
	}

	if unmarshaled.SessionID != tcr.SessionID {
		t.Errorf("SessionID: expected %s, got %s", tcr.SessionID, unmarshaled.SessionID)
	}
	if unmarshaled.Type != tcr.Type {
		t.Errorf("Type: expected %s, got %s", tcr.Type, unmarshaled.Type)
	}
	if unmarshaled.Name != tcr.Name {
		t.Errorf("Name: expected %s, got %s", tcr.Name, unmarshaled.Name)
	}
	if unmarshaled.DurationMs != tcr.DurationMs {
		t.Errorf("DurationMs: expected %d, got %d", tcr.DurationMs, unmarshaled.DurationMs)
	}
	if unmarshaled.Success != tcr.Success {
		t.Errorf("Success: expected %v, got %v", tcr.Success, unmarshaled.Success)
	}
	if unmarshaled.CalledAt != tcr.CalledAt {
		t.Errorf("CalledAt: expected %s, got %s", tcr.CalledAt, unmarshaled.CalledAt)
	}

	t.Logf("ToolCallRecord JSON: %s", string(data))
}

func TestReflectionLogJSON(t *testing.T) {
	rl := ReflectionLog{
		TriggerType:      "MANUAL",
		TriggerDetail:    "user-triggered analysis",
		SessionsAnalyzed: 15,
		ReflectorTokens:  75000,
		DurationMs:       2345,
		ReportPath:       "/reports/2024-01-01 reflection",
		Status:           "SUCCESS",
		TriggeredAt:      "2024-01-01T12:00:00Z",
	}

	data, err := json.Marshal(rl)
	if err != nil {
		t.Fatalf("Failed to marshal ReflectionLog: %v", err)
	}

	var unmarshaled ReflectionLog
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ReflectionLog: %v", err)
	}

	if unmarshaled.TriggerType != rl.TriggerType {
		t.Errorf("TriggerType: expected %s, got %s", rl.TriggerType, unmarshaled.TriggerType)
	}
	if unmarshaled.TriggerDetail != rl.TriggerDetail {
		t.Errorf("TriggerDetail: expected %s, got %s", rl.TriggerDetail, unmarshaled.TriggerDetail)
	}
	if unmarshaled.SessionsAnalyzed != rl.SessionsAnalyzed {
		t.Errorf("SessionsAnalyzed: expected %d, got %d", rl.SessionsAnalyzed, unmarshaled.SessionsAnalyzed)
	}
	if unmarshaled.ReflectorTokens != rl.ReflectorTokens {
		t.Errorf("ReflectorTokens: expected %d, got %d", rl.ReflectorTokens, unmarshaled.ReflectorTokens)
	}
	if unmarshaled.DurationMs != rl.DurationMs {
		t.Errorf("DurationMs: expected %d, got %d", rl.DurationMs, unmarshaled.DurationMs)
	}
	if unmarshaled.ReportPath != rl.ReportPath {
		t.Errorf("ReportPath: expected %s, got %s", rl.ReportPath, unmarshaled.ReportPath)
	}
	if unmarshaled.Status != rl.Status {
		t.Errorf("Status: expected %s, got %s", rl.Status, unmarshaled.Status)
	}
	if unmarshaled.TriggeredAt != rl.TriggeredAt {
		t.Errorf("TriggeredAt: expected %s, got %s", rl.TriggeredAt, unmarshaled.TriggeredAt)
	}

	t.Logf("ReflectionLog JSON: %s", string(data))
}

func TestWatermarkJSON(t *testing.T) {
	wm := Watermark{
		ToolType:        "opencode",
		LastReflectedAt: "2024-01-01T12:00:00Z",
	}

	data, err := json.Marshal(wm)
	if err != nil {
		t.Fatalf("Failed to marshal Watermark: %v", err)
	}

	var unmarshaled Watermark
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Watermark: %v", err)
	}

	if unmarshaled.ToolType != wm.ToolType {
		t.Errorf("ToolType: expected %s, got %s", wm.ToolType, unmarshaled.ToolType)
	}
	if unmarshaled.LastReflectedAt != wm.LastReflectedAt {
		t.Errorf("LastReflectedAt: expected %s, got %s", wm.LastReflectedAt, unmarshaled.LastReflectedAt)
	}

	t.Logf("Watermark JSON: %s", string(data))
}
