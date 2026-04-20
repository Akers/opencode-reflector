package engine

import (
	"encoding/json"
	"testing"
)

func TestSentimentResultJSON(t *testing.T) {
	result := SentimentResult{
		NegativeRatio: 0.2,
		AttitudeScore: 7.5,
		ApprovalRatio: 0.8,
	}

	// Marshal to JSON
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var parsed SentimentResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify values
	if parsed.NegativeRatio != 0.2 {
		t.Errorf("expected NegativeRatio 0.2, got %v", parsed.NegativeRatio)
	}
	if parsed.AttitudeScore != 7.5 {
		t.Errorf("expected AttitudeScore 7.5, got %v", parsed.AttitudeScore)
	}
	if parsed.ApprovalRatio != 0.8 {
		t.Errorf("expected ApprovalRatio 0.8, got %v", parsed.ApprovalRatio)
	}
}

func TestSanitizeMessages(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "API key redaction",
			input:    []string{"sk-abc123def456ghi789jkl012mno345"},
			expected: []string{"[REDACTED_API_KEY]"},
		},
		{
			name:     "Bearer token redaction",
			input:    []string{"Bearer eyJhbGciOiJIUzI1NiJ9"},
			expected: []string{"Bearer [REDACTED]"},
		},
		{
			name:     "Password in URL redaction",
			input:    []string{"mysql://user:secret123@localhost:3306/db"},
			expected: []string{"mysql://[REDACTED]:[REDACTED]@localhost:3306/db"},
		},
		{
			name:     "Generic password redaction",
			input:    []string{"password=secret123"},
			expected: []string{"password=[REDACTED]"},
		},
		{
			name:     "Normal text unchanged",
			input:    []string{"Hello, how are you?"},
			expected: []string{"Hello, how are you?"},
		},
		{
			name:     "Multiple messages",
			input:    []string{"sk-abc123def456ghi789jkl012mno345", "Bearer eyJhbGciOiJIUzI1NiJ9", "normal text"},
			expected: []string{"[REDACTED_API_KEY]", "Bearer [REDACTED]", "normal text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeMessages(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d messages, got %d", len(tt.expected), len(result))
			}
			for i, msg := range result {
				if msg != tt.expected[i] {
					t.Errorf("message %d: expected %q, got %q", i, tt.expected[i], msg)
				}
			}
		})
	}
}
