package engine

import (
	"encoding/json"
	"fmt"

	"github.com/akers/opencode-reflector/internal/model"
)

// ParseSession parses a raw JSON message into a CanonicalSession
func ParseSession(raw json.RawMessage) (*model.CanonicalSession, error) {
	if raw == nil {
		return nil, fmt.Errorf("invalid JSON: nil input")
	}

	// Try to detect if it's a JSON object (starts with {)
	trimmed := []byte{}
	for i := 0; i < len(raw); i++ {
		if raw[i] == ' ' || raw[i] == '\t' || raw[i] == '\n' || raw[i] == '\r' {
			continue
		}
		trimmed = raw[i:]
		break
	}

	if len(trimmed) == 0 {
		return nil, fmt.Errorf("invalid JSON: empty input")
	}

	if trimmed[0] != '{' {
		return nil, fmt.Errorf("invalid JSON: expected object, got %c", trimmed[0])
	}

	var session model.CanonicalSession
	if err := json.Unmarshal(raw, &session); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return &session, nil
}
