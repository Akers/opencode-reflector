package model

import "encoding/json"

// AgentToolType represents the type of agent tool being used
type AgentToolType string

const (
	AgentToolTypeOpencode    AgentToolType = "opencode"
	AgentToolTypeOpenclaw    AgentToolType = "openclaw"
	AgentToolTypeClaudecode  AgentToolType = "claudecode"
)

// MessageRole represents the role of a message sender
type MessageRole string

const (
	MessageRoleHuman   MessageRole = "human"
	MessageRoleAgent   MessageRole = "agent"
	MessageRoleSystem  MessageRole = "system"
)

// ToolCallType represents the type of tool call
type ToolCallType string

const (
	ToolCallTypeTool   ToolCallType = "TOOL"
	ToolCallTypeMCP    ToolCallType = "MCP"
	ToolCallTypeSkill  ToolCallType = "SKILL"
)

// CanonicalMessage represents a message in a canonical session
type CanonicalMessage struct {
	Role             MessageRole            `json:"role"`
	Content          string                 `json:"content"`
	Timestamp        string                 `json:"timestamp"`
	AgentName        *string                `json:"agent_name,omitempty"`
	PromptTokens     *int                   `json:"prompt_tokens,omitempty"`
	CompletionTokens *int                   `json:"completion_tokens,omitempty"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// CanonicalToolCall represents a tool call in a canonical session
type CanonicalToolCall struct {
	Type       ToolCallType `json:"type"`
	Name       string       `json:"name"`
	DurationMs *int64       `json:"duration_ms,omitempty"`
	Success    bool         `json:"success"`
	CalledAt   string       `json:"called_at"`
}

// CanonicalSession represents a complete coding session
type CanonicalSession struct {
	ID        string              `json:"id"`
	ToolType  AgentToolType       `json:"tool_type"`
	Title     *string             `json:"title,omitempty"`
	Messages  []CanonicalMessage `json:"messages"`
	ToolCalls []CanonicalToolCall `json:"tool_calls"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// CapabilityMatrix represents the capabilities of a session
type CapabilityMatrix struct {
	TokenMetrics      bool `json:"token_metrics"`
	ToolCallDetails   bool `json:"tool_call_details"`
	MCPCallDetails    bool `json:"mcp_call_details"`
	SkillCallDetails  bool `json:"skill_call_details"`
	AgentNames        bool `json:"agent_names"`
}

// MarshalJSON implements custom marshaling for CanonicalMessage
func (m CanonicalMessage) MarshalJSON() ([]byte, error) {
	type Alias CanonicalMessage
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(m),
	})
}

// UnmarshalJSON implements custom unmarshaling for CanonicalMessage
func (m *CanonicalMessage) UnmarshalJSON(data []byte) error {
	type Alias CanonicalMessage
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	return json.Unmarshal(data, aux)
}

// MarshalJSON implements custom marshaling for CanonicalToolCall
func (t CanonicalToolCall) MarshalJSON() ([]byte, error) {
	type Alias CanonicalToolCall
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(t),
	})
}

// UnmarshalJSON implements custom unmarshaling for CanonicalToolCall
func (t *CanonicalToolCall) UnmarshalJSON(data []byte) error {
	type Alias CanonicalToolCall
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	return json.Unmarshal(data, aux)
}

// MarshalJSON implements custom marshaling for CanonicalSession
func (s CanonicalSession) MarshalJSON() ([]byte, error) {
	type Alias CanonicalSession
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(s),
	})
}

// UnmarshalJSON implements custom unmarshaling for CanonicalSession
func (s *CanonicalSession) UnmarshalJSON(data []byte) error {
	type Alias CanonicalSession
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	return json.Unmarshal(data, aux)
}
