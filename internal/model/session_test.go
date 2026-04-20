package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCanonicalMessageJSONRoundTrip(t *testing.T) {
	agentName := "main"
	promptTokens := 1500
	completionTokens := 800
	metadata := map[string]interface{}{"key": "value"}

	msg := CanonicalMessage{
		Role:             MessageRoleAgent,
		Content:          "test",
		Timestamp:        "2026-04-19T09:30:00Z",
		AgentName:        &agentName,
		PromptTokens:     &promptTokens,
		CompletionTokens: &completionTokens,
		Metadata:         metadata,
	}

	// Marshal
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal CanonicalMessage: %v", err)
	}

	// Unmarshal
	var decoded CanonicalMessage
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal CanonicalMessage: %v", err)
	}

	// Verify all fields
	if decoded.Role != msg.Role {
		t.Errorf("Role mismatch: got %v, want %v", decoded.Role, msg.Role)
	}
	if decoded.Content != msg.Content {
		t.Errorf("Content mismatch: got %v, want %v", decoded.Content, msg.Content)
	}
	if decoded.Timestamp != msg.Timestamp {
		t.Errorf("Timestamp mismatch: got %v, want %v", decoded.Timestamp, msg.Timestamp)
	}
	if decoded.AgentName == nil || *decoded.AgentName != agentName {
		t.Errorf("AgentName mismatch: got %v, want %v", decoded.AgentName, agentName)
	}
	if decoded.PromptTokens == nil || *decoded.PromptTokens != promptTokens {
		t.Errorf("PromptTokens mismatch: got %v, want %v", decoded.PromptTokens, promptTokens)
	}
	if decoded.CompletionTokens == nil || *decoded.CompletionTokens != completionTokens {
		t.Errorf("CompletionTokens mismatch: got %v, want %v", decoded.CompletionTokens, completionTokens)
	}
	if decoded.Metadata["key"] != "value" {
		t.Errorf("Metadata mismatch: got %v, want %v", decoded.Metadata, metadata)
	}
}

func TestCanonicalMessageOptionalFields(t *testing.T) {
	// Test with all optional fields nil
	msg := CanonicalMessage{
		Role:      MessageRoleHuman,
		Content:   "hello",
		Timestamp: time.Now().Format(time.RFC3339),
		Metadata:  make(map[string]interface{}),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal CanonicalMessage with nil optionals: %v", err)
	}

	var decoded CanonicalMessage
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal CanonicalMessage with nil optionals: %v", err)
	}

	if decoded.AgentName != nil {
		t.Errorf("AgentName should be nil, got %v", decoded.AgentName)
	}
	if decoded.PromptTokens != nil {
		t.Errorf("PromptTokens should be nil, got %v", decoded.PromptTokens)
	}
	if decoded.CompletionTokens != nil {
		t.Errorf("CompletionTokens should be nil, got %v", decoded.CompletionTokens)
	}

	// Verify the JSON doesn't contain these fields when nil
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	if _, exists := raw["agent_name"]; exists {
		t.Error("agent_name should not be present in JSON when nil")
	}
	if _, exists := raw["prompt_tokens"]; exists {
		t.Error("prompt_tokens should not be present in JSON when nil")
	}
	if _, exists := raw["completion_tokens"]; exists {
		t.Error("completion_tokens should not be present in JSON when nil")
	}
}

func TestCanonicalToolCallJSONRoundTrip(t *testing.T) {
	durationMs := int64(150)
	toolCall := CanonicalToolCall{
		Type:       ToolCallTypeMCP,
		Name:       "filesystem/read",
		DurationMs: &durationMs,
		Success:    true,
		CalledAt:   "2026-04-19T09:30:00Z",
	}

	// Marshal
	data, err := json.Marshal(toolCall)
	if err != nil {
		t.Fatalf("Failed to marshal CanonicalToolCall: %v", err)
	}

	// Unmarshal
	var decoded CanonicalToolCall
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal CanonicalToolCall: %v", err)
	}

	// Verify all fields
	if decoded.Type != toolCall.Type {
		t.Errorf("Type mismatch: got %v, want %v", decoded.Type, toolCall.Type)
	}
	if decoded.Name != toolCall.Name {
		t.Errorf("Name mismatch: got %v, want %v", decoded.Name, toolCall.Name)
	}
	if decoded.DurationMs == nil || *decoded.DurationMs != durationMs {
		t.Errorf("DurationMs mismatch: got %v, want %v", decoded.DurationMs, durationMs)
	}
	if decoded.Success != toolCall.Success {
		t.Errorf("Success mismatch: got %v, want %v", decoded.Success, toolCall.Success)
	}
	if decoded.CalledAt != toolCall.CalledAt {
		t.Errorf("CalledAt mismatch: got %v, want %v", decoded.CalledAt, toolCall.CalledAt)
	}
}

func TestCanonicalSessionJSONRoundTrip(t *testing.T) {
	title := "Test Session"
	agentName := "main"
	promptTokens := 100
	completionTokens := 50
	durationMs := int64(200)

	messages := []CanonicalMessage{
		{
			Role:      MessageRoleHuman,
			Content:   "Hello, can you help me?",
			Timestamp: "2026-04-19T09:00:00Z",
			Metadata:  map[string]interface{}{},
		},
		{
			Role:             MessageRoleAgent,
			Content:          "Of course!",
			Timestamp:        "2026-04-19T09:01:00Z",
			AgentName:        &agentName,
			PromptTokens:     &promptTokens,
			CompletionTokens: &completionTokens,
			Metadata:         map[string]interface{}{"language": "en"},
		},
	}

	toolCalls := []CanonicalToolCall{
		{
			Type:       ToolCallTypeTool,
			Name:       "bash/execute",
			DurationMs: &durationMs,
			Success:    true,
			CalledAt:   "2026-04-19T09:00:30Z",
		},
	}

	session := CanonicalSession{
		ID:        "session-123",
		ToolType:  AgentToolTypeOpencode,
		Title:     &title,
		Messages:  messages,
		ToolCalls: toolCalls,
		Metadata:  map[string]interface{}{"version": "1.0"},
	}

	// Marshal
	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("Failed to marshal CanonicalSession: %v", err)
	}

	// Unmarshal
	var decoded CanonicalSession
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal CanonicalSession: %v", err)
	}

	// Verify top-level fields
	if decoded.ID != session.ID {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID, session.ID)
	}
	if decoded.ToolType != session.ToolType {
		t.Errorf("ToolType mismatch: got %v, want %v", decoded.ToolType, session.ToolType)
	}
	if decoded.Title == nil || *decoded.Title != title {
		t.Errorf("Title mismatch: got %v, want %v", decoded.Title, title)
	}

	// Verify messages
	if len(decoded.Messages) != 2 {
		t.Fatalf("Messages length mismatch: got %d, want 2", len(decoded.Messages))
	}
	if decoded.Messages[0].Role != MessageRoleHuman {
		t.Errorf("Messages[0].Role mismatch: got %v, want %v", decoded.Messages[0].Role, MessageRoleHuman)
	}
	if decoded.Messages[1].Role != MessageRoleAgent {
		t.Errorf("Messages[1].Role mismatch: got %v, want %v", decoded.Messages[1].Role, MessageRoleAgent)
	}
	if decoded.Messages[1].AgentName == nil || *decoded.Messages[1].AgentName != agentName {
		t.Errorf("Messages[1].AgentName mismatch: got %v, want %v", decoded.Messages[1].AgentName, agentName)
	}

	// Verify tool calls
	if len(decoded.ToolCalls) != 1 {
		t.Fatalf("ToolCalls length mismatch: got %d, want 1", len(decoded.ToolCalls))
	}
	if decoded.ToolCalls[0].Type != ToolCallTypeTool {
		t.Errorf("ToolCalls[0].Type mismatch: got %v, want %v", decoded.ToolCalls[0].Type, ToolCallTypeTool)
	}
	if decoded.ToolCalls[0].Name != "bash/execute" {
		t.Errorf("ToolCalls[0].Name mismatch: got %v, want %v", decoded.ToolCalls[0].Name, "bash/execute")
	}
}

func TestCapabilityMatrixDefaults(t *testing.T) {
	matrix := CapabilityMatrix{}

	if matrix.TokenMetrics != false {
		t.Errorf("TokenMetrics should be false by default, got %v", matrix.TokenMetrics)
	}
	if matrix.ToolCallDetails != false {
		t.Errorf("ToolCallDetails should be false by default, got %v", matrix.ToolCallDetails)
	}
	if matrix.MCPCallDetails != false {
		t.Errorf("MCPCallDetails should be false by default, got %v", matrix.MCPCallDetails)
	}
	if matrix.SkillCallDetails != false {
		t.Errorf("SkillCallDetails should be false by default, got %v", matrix.SkillCallDetails)
	}
	if matrix.AgentNames != false {
		t.Errorf("AgentNames should be false by default, got %v", matrix.AgentNames)
	}

	// Also verify JSON marshaling gives all false
	data, err := json.Marshal(matrix)
	if err != nil {
		t.Fatalf("Failed to marshal CapabilityMatrix: %v", err)
	}

	var decoded CapabilityMatrix
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal CapabilityMatrix: %v", err)
	}

	if decoded.TokenMetrics != false {
		t.Errorf("Unmarshaled TokenMetrics should be false, got %v", decoded.TokenMetrics)
	}
	if decoded.ToolCallDetails != false {
		t.Errorf("Unmarshaled ToolCallDetails should be false, got %v", decoded.ToolCallDetails)
	}
	if decoded.MCPCallDetails != false {
		t.Errorf("Unmarshaled MCPCallDetails should be false, got %v", decoded.MCPCallDetails)
	}
	if decoded.SkillCallDetails != false {
		t.Errorf("Unmarshaled SkillCallDetails should be false, got %v", decoded.SkillCallDetails)
	}
	if decoded.AgentNames != false {
		t.Errorf("Unmarshaled AgentNames should be false, got %v", decoded.AgentNames)
	}
}

func TestMessageRoleValues(t *testing.T) {
	roles := []MessageRole{MessageRoleHuman, MessageRoleAgent, MessageRoleSystem}
	expected := []string{"human", "agent", "system"}

	for i, role := range roles {
		if string(role) != expected[i] {
			t.Errorf("Role mismatch: got %v, want %v", role, expected[i])
		}
	}
}

func TestToolCallTypeValues(t *testing.T) {
	types := []ToolCallType{ToolCallTypeTool, ToolCallTypeMCP, ToolCallTypeSkill}
	expected := []string{"TOOL", "MCP", "SKILL"}

	for i, typ := range types {
		if string(typ) != expected[i] {
			t.Errorf("Type mismatch: got %v, want %v", typ, expected[i])
		}
	}
}

func TestAgentToolTypeValues(t *testing.T) {
	types := []AgentToolType{AgentToolTypeOpencode, AgentToolTypeOpenclaw, AgentToolTypeClaudecode}
	expected := []string{"opencode", "openclaw", "claudecode"}

	for i, typ := range types {
		if string(typ) != expected[i] {
			t.Errorf("Type mismatch: got %v, want %v", typ, expected[i])
		}
	}
}
