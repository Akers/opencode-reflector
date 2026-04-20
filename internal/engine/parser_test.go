package engine

import (
	"encoding/json"
	"testing"
)

func TestParseSession(t *testing.T) {
	raw := json.RawMessage(`{
		"id": "test-session-123",
		"tool_type": "opencode",
		"title": "Test Session",
		"messages": [
			{
				"role": "human",
				"content": "Hello",
				"timestamp": "2024-01-01T10:00:00Z"
			}
		],
		"tool_calls": []
	}`)

	session, err := ParseSession(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if session.ID != "test-session-123" {
		t.Errorf("expected ID 'test-session-123', got '%s'", session.ID)
	}
	if session.ToolType != "opencode" {
		t.Errorf("expected ToolType 'opencode', got '%s'", session.ToolType)
	}
	if session.Title == nil || *session.Title != "Test Session" {
		t.Errorf("expected Title 'Test Session', got %v", session.Title)
	}
	if len(session.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(session.Messages))
	}
}

func TestParseSessionInvalidJSON(t *testing.T) {
	testCases := []struct {
		name string
		raw  json.RawMessage
	}{
		{
			name: "nil input",
			raw:  nil,
		},
		{
			name: "empty input",
			raw:  json.RawMessage(""),
		},
		{
			name: "invalid JSON string",
			raw:  json.RawMessage("not json"),
		},
		{
			name: "array instead of object",
			raw:  json.RawMessage(`[1, 2, 3]`),
		},
		{
			name: "truncated JSON",
			raw:  json.RawMessage(`{"id": "test"`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseSession(tc.raw)
			if err == nil {
				t.Errorf("expected error for %s, got nil", tc.name)
			}
		})
	}
}
