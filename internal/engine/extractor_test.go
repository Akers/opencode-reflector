package engine

import (
	"testing"
	"time"

	"github.com/akers/opencode-reflector/internal/model"
)

func ptrString(s string) *string {
	return &s
}

func ptrInt(i int) *int {
	return &i
}

func ptrInt64(i int64) *int64 {
	return &i
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestExtractTimeMetrics(t *testing.T) {
	session := &model.CanonicalSession{
		ID:       "test-session",
		ToolType: model.AgentToolTypeOpencode,
		Messages: []model.CanonicalMessage{
			{
				Role:      model.MessageRoleHuman,
				Content:   "Start",
				Timestamp: "2024-01-01T10:00:00Z",
			},
			{
				Role:      model.MessageRoleAgent,
				Content:   "Response 1",
				Timestamp: "2024-01-01T10:01:00Z",
			},
			{
				Role:      model.MessageRoleHuman,
				Content:   "End",
				Timestamp: "2024-01-01T10:05:00Z",
			},
		},
	}

	startedAt, endedAt, durationSeconds, humanThinkTimeSeconds, agentThinkTimeSeconds := ExtractTimeMetrics(session)

	if startedAt != "2024-01-01T10:00:00Z" {
		t.Errorf("expected startedAt '2024-01-01T10:00:00Z', got '%s'", startedAt)
	}
	if endedAt != "2024-01-01T10:05:00Z" {
		t.Errorf("expected endedAt '2024-01-01T10:05:00Z', got '%s'", endedAt)
	}
	// 5 minutes = 300 seconds
	if durationSeconds != 300.0 {
		t.Errorf("expected durationSeconds 300.0, got %f", durationSeconds)
	}
	// Human think time: from first human (10:00:00) to second human (10:05:00) = 300s
	if humanThinkTimeSeconds != 300.0 {
		t.Errorf("expected humanThinkTimeSeconds 300.0, got %f", humanThinkTimeSeconds)
	}
	// Agent think time: only one agent message, no interval
	if agentThinkTimeSeconds != 0.0 {
		t.Errorf("expected agentThinkTimeSeconds 0.0, got %f", agentThinkTimeSeconds)
	}
}

func TestExtractTimeMetricsEmpty(t *testing.T) {
	// Empty messages
	session := &model.CanonicalSession{
		ID:       "test-session",
		ToolType: model.AgentToolTypeOpencode,
		Messages: []model.CanonicalMessage{},
	}

	startedAt, endedAt, durationSeconds, humanThinkTimeSeconds, agentThinkTimeSeconds := ExtractTimeMetrics(session)

	if startedAt != "" {
		t.Errorf("expected empty startedAt, got '%s'", startedAt)
	}
	if endedAt != "" {
		t.Errorf("expected empty endedAt, got '%s'", endedAt)
	}
	if durationSeconds != 0.0 {
		t.Errorf("expected 0.0 durationSeconds, got %f", durationSeconds)
	}
	if humanThinkTimeSeconds != 0.0 {
		t.Errorf("expected 0.0 humanThinkTimeSeconds, got %f", humanThinkTimeSeconds)
	}
	if agentThinkTimeSeconds != 0.0 {
		t.Errorf("expected 0.0 agentThinkTimeSeconds, got %f", agentThinkTimeSeconds)
	}

	// Nil session
	startedAt, endedAt, durationSeconds, humanThinkTimeSeconds, agentThinkTimeSeconds = ExtractTimeMetrics(nil)
	if startedAt != "" || endedAt != "" || durationSeconds != 0.0 {
		t.Error("nil session should return all zeros")
	}
}

func TestExtractTokenMetrics(t *testing.T) {
	session := &model.CanonicalSession{
		ID:       "test-session",
		ToolType: model.AgentToolTypeOpencode,
		Messages: []model.CanonicalMessage{
			{
				Role:             model.MessageRoleHuman,
				PromptTokens:     ptrInt(100),
				CompletionTokens: ptrInt(50),
			},
			{
				Role:             model.MessageRoleAgent,
				PromptTokens:     ptrInt(150),
				CompletionTokens: ptrInt(75),
			},
			{
				Role:             model.MessageRoleHuman,
				PromptTokens:     ptrInt(200),
				CompletionTokens: ptrInt(100),
			},
		},
	}

	promptTokens, completionTokens, totalTokens, modelRequestCount := ExtractTokenMetrics(session)

	if promptTokens != 450.0 {
		t.Errorf("expected promptTokens 450, got %f", promptTokens)
	}
	if completionTokens != 225.0 {
		t.Errorf("expected completionTokens 225, got %f", completionTokens)
	}
	if totalTokens != 675.0 {
		t.Errorf("expected totalTokens 675, got %f", totalTokens)
	}
	if modelRequestCount != 3.0 {
		t.Errorf("expected modelRequestCount 3, got %f", modelRequestCount)
	}
}

func TestExtractTokenMetricsNA(t *testing.T) {
	session := &model.CanonicalSession{
		ID:       "test-session",
		ToolType: model.AgentToolTypeOpencode,
		Messages: []model.CanonicalMessage{
			{
				Role:    model.MessageRoleHuman,
				Content: "No token data here",
			},
			{
				Role:    model.MessageRoleAgent,
				Content: "No token data here either",
			},
		},
	}

	promptTokens, completionTokens, totalTokens, modelRequestCount := ExtractTokenMetrics(session)

	if promptTokens != -1.0 {
		t.Errorf("expected promptTokens -1, got %f", promptTokens)
	}
	if completionTokens != -1.0 {
		t.Errorf("expected completionTokens -1, got %f", completionTokens)
	}
	if totalTokens != -1.0 {
		t.Errorf("expected totalTokens -1, got %f", totalTokens)
	}
	if modelRequestCount != -1.0 {
		t.Errorf("expected modelRequestCount -1, got %f", modelRequestCount)
	}
}

func TestExtractToolMetrics(t *testing.T) {
	session := &model.CanonicalSession{
		ID:       "test-session",
		ToolType: model.AgentToolTypeOpencode,
		ToolCalls: []model.CanonicalToolCall{
			{
				Type:       model.ToolCallTypeTool,
				Name:       "bash",
				DurationMs: ptrInt64(100),
				Success:    true,
				CalledAt:   "2024-01-01T10:00:00Z",
			},
			{
				Type:       model.ToolCallTypeTool,
				Name:       "read",
				DurationMs: ptrInt64(200),
				Success:    true,
				CalledAt:   "2024-01-01T10:00:01Z",
			},
			{
				Type:       model.ToolCallTypeTool,
				Name:       "write",
				DurationMs: ptrInt64(150),
				Success:    false,
				CalledAt:   "2024-01-01T10:00:02Z",
			},
			{
				Type:       model.ToolCallTypeMCP,
				Name:       "mcp_tool",
				DurationMs: ptrInt64(50),
				Success:    true,
				CalledAt:   "2024-01-01T10:00:03Z",
			},
			{
				Type:       model.ToolCallTypeMCP,
				Name:       "mcp_tool_2",
				DurationMs: ptrInt64(75),
				Success:    true,
				CalledAt:   "2024-01-01T10:00:04Z",
			},
			{
				Type:       model.ToolCallTypeSkill,
				Name:       "skill_call",
				Success:    true,
				CalledAt:   "2024-01-01T10:00:05Z",
			},
		},
	}

	toolCallCount, toolSuccessCount, mcpCallCount, skillCallCount, toolAvgDurationMs, mcpAvgDurationMs, records := ExtractToolMetrics(session)

	if toolCallCount != 3.0 {
		t.Errorf("expected toolCallCount 3, got %f", toolCallCount)
	}
	if toolSuccessCount != 2.0 {
		t.Errorf("expected toolSuccessCount 2, got %f", toolSuccessCount)
	}
	if mcpCallCount != 2.0 {
		t.Errorf("expected mcpCallCount 2, got %f", mcpCallCount)
	}
	if skillCallCount != 1.0 {
		t.Errorf("expected skillCallCount 1, got %f", skillCallCount)
	}
	// (100 + 200 + 150) / 3 = 450 / 3 = 150
	if toolAvgDurationMs != 150.0 {
		t.Errorf("expected toolAvgDurationMs 150, got %f", toolAvgDurationMs)
	}
	// (50 + 75) / 2 = 125 / 2 = 62.5
	if mcpAvgDurationMs != 62.5 {
		t.Errorf("expected mcpAvgDurationMs 62.5, got %f", mcpAvgDurationMs)
	}
	if len(records) != 6 {
		t.Errorf("expected 6 records, got %d", len(records))
	}
}

func TestExtractAgentMetrics(t *testing.T) {
	agent1 := "agent-alpha"
	agent2 := "agent-beta"
	session := &model.CanonicalSession{
		ID:       "test-session",
		ToolType: model.AgentToolTypeOpencode,
		Messages: []model.CanonicalMessage{
			{
				Role:      model.MessageRoleAgent,
				Content:   "Hello from alpha",
				AgentName: &agent1,
			},
			{
				Role:      model.MessageRoleAgent,
				Content:   "Alpha again",
				AgentName: &agent1,
			},
			{
				Role:      model.MessageRoleAgent,
				Content:   "Hello from beta",
				AgentName: &agent2,
			},
			{
				Role:      model.MessageRoleHuman,
				Content:   "Human message",
			},
			{
				Role:      model.MessageRoleAgent,
				Content:   "Beta again",
				AgentName: &agent2,
			},
		},
	}

	participationCount, participations, agentNames := ExtractAgentMetrics(session)

	if participationCount != 2.0 {
		t.Errorf("expected participationCount 2, got %f", participationCount)
	}
	if len(participations) != 2 {
		t.Errorf("expected 2 participations, got %d", len(participations))
	}
	// Check agent names (order may vary due to map iteration)
	expectedNames := map[string]bool{"agent-alpha": true, "agent-beta": true}
	for _, name := range []string{agentNames} {
		delete(expectedNames, name)
	}
	if agentNames != "agent-alpha,agent-beta" && agentNames != "agent-beta,agent-alpha" {
		t.Errorf("expected agent names 'agent-alpha,agent-beta' or reverse, got '%s'", agentNames)
	}
}

func TestExtractAgentMetricsNoAgents(t *testing.T) {
	session := &model.CanonicalSession{
		ID:       "test-session",
		ToolType: model.AgentToolTypeOpencode,
		Messages: []model.CanonicalMessage{
			{
				Role:    model.MessageRoleHuman,
				Content: "Human message",
			},
		},
	}

	participationCount, participations, agentNames := ExtractAgentMetrics(session)

	if participationCount != 0.0 {
		t.Errorf("expected participationCount 0, got %f", participationCount)
	}
	if len(participations) != 0 {
		t.Errorf("expected 0 participations, got %d", len(participations))
	}
	if agentNames != "" {
		t.Errorf("expected empty agentNames, got '%s'", agentNames)
	}
}

func TestExtractMessageMetrics(t *testing.T) {
	session := &model.CanonicalSession{
		ID:       "test-session",
		ToolType: model.AgentToolTypeOpencode,
		Messages: []model.CanonicalMessage{
			{Role: model.MessageRoleSystem},
			{Role: model.MessageRoleHuman},
			{Role: model.MessageRoleAgent},
			{Role: model.MessageRoleHuman},
			{Role: model.MessageRoleAgent},
			{Role: model.MessageRoleHuman},
			{Role: model.MessageRoleSystem},
			{Role: model.MessageRoleAgent},
			{Role: model.MessageRoleHuman},
			{Role: model.MessageRoleAgent},
			{Role: model.MessageRoleHuman},
			{Role: model.MessageRoleAgent},
			{Role: model.MessageRoleHuman},
			{Role: model.MessageRoleAgent},
			{Role: model.MessageRoleAgent},
		},
	}

	total, agentCount, humanCount := ExtractMessageMetrics(session)

	// 6 human + 7 agent = 13 total (system not counted)
	if humanCount != 6.0 {
		t.Errorf("expected humanCount 6, got %f", humanCount)
	}
	if agentCount != 7.0 {
		t.Errorf("expected agentCount 7, got %f", agentCount)
	}
	if total != 13.0 {
		t.Errorf("expected total 13, got %f", total)
	}
}

func TestExtractHumanParticipation(t *testing.T) {
	session := &model.CanonicalSession{
		ID:       "test-session",
		ToolType: model.AgentToolTypeOpencode,
		Messages: []model.CanonicalMessage{
			{
				Role:      model.MessageRoleHuman,
				Content:   "Short",
				Timestamp: "2024-01-01T10:00:00Z",
			},
			{
				Role:      model.MessageRoleAgent,
				Content:   "Response",
				Timestamp: "2024-01-01T10:00:01Z",
			},
			{
				Role:      model.MessageRoleHuman,
				Content:   "Medium length message",
				Timestamp: "2024-01-01T10:01:00Z",
			},
			{
				Role:      model.MessageRoleAgent,
				Content:   "Another response",
				Timestamp: "2024-01-01T10:01:01Z",
			},
			{
				Role:      model.MessageRoleHuman,
				Content:   "A much longer human message here",
				Timestamp: "2024-01-01T10:05:00Z",
			},
		},
	}

	ratio, avgIntervalSeconds, avgChars := ExtractHumanParticipation(session)

	// ratio = 3 / (3 + 2) = 3/5 = 0.6
	if ratio != 0.6 {
		t.Errorf("expected ratio 0.6, got %f", ratio)
	}
	// avg interval: ((1min - 0) + (5min - 1min)) / 2 = (60 + 240) / 2 = 150
	if avgIntervalSeconds != 150.0 {
		t.Errorf("expected avgIntervalSeconds 150, got %f", avgIntervalSeconds)
	}
	// avg chars: 5 + 21 + 32 = 58 / 3 = 19.333333...
	avgCharsExpected := 58.0 / 3.0
	if abs(avgChars-avgCharsExpected) > 0.001 {
		t.Errorf("expected avgChars ~19.333, got %f", avgChars)
	}
}

func TestExtractAll(t *testing.T) {
	agentName := "test-agent"
	title := "Test Session"
	session := &model.CanonicalSession{
		ID:        "full-test-session",
		ToolType:  model.AgentToolTypeOpencode,
		Title:     &title,
		Messages: []model.CanonicalMessage{
			{
				Role:             model.MessageRoleHuman,
				Content:          "Hello agent",
				Timestamp:        "2024-01-01T10:00:00Z",
				PromptTokens:     ptrInt(100),
				CompletionTokens: ptrInt(50),
			},
			{
				Role:             model.MessageRoleAgent,
				Content:          "Hello human",
				Timestamp:        "2024-01-01T10:00:30Z",
				AgentName:        &agentName,
				PromptTokens:     ptrInt(150),
				CompletionTokens: ptrInt(75),
			},
			{
				Role:             model.MessageRoleHuman,
				Content:          "Goodbye",
				Timestamp:        "2024-01-01T10:01:00Z",
				PromptTokens:     ptrInt(100),
				CompletionTokens: ptrInt(50),
			},
		},
		ToolCalls: []model.CanonicalToolCall{
			{
				Type:       model.ToolCallTypeTool,
				Name:       "bash",
				DurationMs: ptrInt64(100),
				Success:    true,
				CalledAt:   "2024-01-01T10:00:15Z",
			},
		},
	}

	metrics := ExtractAll(session)

	if metrics.SessionID != "full-test-session" {
		t.Errorf("expected SessionID 'full-test-session', got '%s'", metrics.SessionID)
	}
	if metrics.ToolType != "opencode" {
		t.Errorf("expected ToolType 'opencode', got '%s'", metrics.ToolType)
	}
	if metrics.Title == nil || *metrics.Title != "Test Session" {
		t.Errorf("expected Title 'Test Session', got %v", metrics.Title)
	}
	if metrics.StartedAt != "2024-01-01T10:00:00Z" {
		t.Errorf("expected StartedAt '2024-01-01T10:00:00Z', got '%s'", metrics.StartedAt)
	}
	if metrics.EndedAt != "2024-01-01T10:01:00Z" {
		t.Errorf("expected EndedAt '2024-01-01T10:01:00Z', got '%s'", metrics.EndedAt)
	}
	if metrics.DurationSeconds != 60.0 {
		t.Errorf("expected DurationSeconds 60, got %f", metrics.DurationSeconds)
	}
	if metrics.TotalPromptTokens != 350.0 {
		t.Errorf("expected TotalPromptTokens 350, got %f", metrics.TotalPromptTokens)
	}
	if metrics.TotalCompletionTokens != 175.0 {
		t.Errorf("expected TotalCompletionTokens 175, got %f", metrics.TotalCompletionTokens)
	}
	if metrics.TotalTokens != 525.0 {
		t.Errorf("expected TotalTokens 525, got %f", metrics.TotalTokens)
	}
	if metrics.ModelRequestCount != 3.0 {
		t.Errorf("expected ModelRequestCount 3, got %f", metrics.ModelRequestCount)
	}
	if metrics.ToolCallCount != 1.0 {
		t.Errorf("expected ToolCallCount 1, got %f", metrics.ToolCallCount)
	}
	if metrics.ToolSuccessCount != 1.0 {
		t.Errorf("expected ToolSuccessCount 1, got %f", metrics.ToolSuccessCount)
	}
	if metrics.MCPCallCount != 0.0 {
		t.Errorf("expected MCPCallCount 0, got %f", metrics.MCPCallCount)
	}
	if metrics.SkillCallCount != 0.0 {
		t.Errorf("expected SkillCallCount 0, got %f", metrics.SkillCallCount)
	}
	if metrics.ToolAvgDurationMs != 100.0 {
		t.Errorf("expected ToolAvgDurationMs 100, got %f", metrics.ToolAvgDurationMs)
	}
	if metrics.MCPAvgDurationMs != -1.0 {
		t.Errorf("expected MCPAvgDurationMs -1, got %f", metrics.MCPAvgDurationMs)
	}
	if metrics.AgentParticipationCount != 1.0 {
		t.Errorf("expected AgentParticipationCount 1, got %f", metrics.AgentParticipationCount)
	}
	if metrics.AgentNames != "test-agent" {
		t.Errorf("expected AgentNames 'test-agent', got '%s'", metrics.AgentNames)
	}
	if metrics.TotalMessageCount != 3.0 {
		t.Errorf("expected TotalMessageCount 3, got %f", metrics.TotalMessageCount)
	}
	if metrics.HumanMessageCount != 2.0 {
		t.Errorf("expected HumanMessageCount 2, got %f", metrics.HumanMessageCount)
	}
	if metrics.AgentMessageCount != 1.0 {
		t.Errorf("expected AgentMessageCount 1, got %f", metrics.AgentMessageCount)
	}
	if metrics.TaskStatus != "" {
		t.Errorf("expected empty TaskStatus, got '%s'", metrics.TaskStatus)
	}
	// Check that AnalyzedAt is a valid RFC3339 timestamp
	if _, err := time.Parse(time.RFC3339, metrics.AnalyzedAt); err != nil {
		t.Errorf("expected valid RFC3339 AnalyzedAt, got '%s': %v", metrics.AnalyzedAt, err)
	}
}

func TestExtractAllEmpty(t *testing.T) {
	session := &model.CanonicalSession{
		ID:        "empty-session",
		ToolType:  model.AgentToolTypeOpencode,
		Messages:  []model.CanonicalMessage{},
		ToolCalls: []model.CanonicalToolCall{},
	}

	// Should not panic
	metrics := ExtractAll(session)

	if metrics.SessionID != "empty-session" {
		t.Errorf("expected SessionID 'empty-session', got '%s'", metrics.SessionID)
	}
	if metrics.TotalPromptTokens != -1.0 {
		t.Errorf("expected TotalPromptTokens -1, got %f", metrics.TotalPromptTokens)
	}
	if metrics.ToolCallCount != 0.0 {
		t.Errorf("expected ToolCallCount 0, got %f", metrics.ToolCallCount)
	}
	if metrics.AgentParticipationCount != 0.0 {
		t.Errorf("expected AgentParticipationCount 0, got %f", metrics.AgentParticipationCount)
	}
	if metrics.TotalMessageCount != 0.0 {
		t.Errorf("expected TotalMessageCount 0, got %f", metrics.TotalMessageCount)
	}
}
