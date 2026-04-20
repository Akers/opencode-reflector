package model

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusCompleted   TaskStatus = "COMPLETED"
	TaskStatusInterrupted TaskStatus = "INTERRUPTED"
	TaskStatusUncertain   TaskStatus = "UNCERTAIN"
	TaskStatusAbandoned   TaskStatus = "ABANDONED"
)

// TriggerType represents the type of trigger for reflection
type TriggerType string

const (
	TriggerTypeTime    TriggerType = "TIME"
	TriggerTypeManual  TriggerType = "MANUAL"
	TriggerTypeEvents  TriggerType = "EVENTS"
)

// SessionMetrics contains all metrics for a session
type SessionMetrics struct {
	SessionID string `json:"session_id"`
	ToolType  string `json:"tool_type"`
	Title     *string `json:"title,omitempty"`

	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`

	// Time metrics (M-001~004)
	DurationSeconds         float64 `json:"duration_seconds"`
	ActiveTimeSeconds       float64 `json:"active_time_seconds"`
	HumanThinkTimeSeconds   float64 `json:"human_think_time_seconds"`
	AgentThinkTimeSeconds   float64 `json:"agent_think_time_seconds"`

	// Token metrics (M-005~008)
	TotalPromptTokens     float64 `json:"total_prompt_tokens"`
	TotalCompletionTokens float64 `json:"total_completion_tokens"`
	TotalTokens           float64 `json:"total_tokens"`
	ModelRequestCount     float64 `json:"model_request_count"`

	// Tool metrics (M-009~014)
	ToolCallCount       float64 `json:"tool_call_count"`
	ToolSuccessCount    float64 `json:"tool_success_count"`
	MCPCallCount        float64 `json:"mcp_call_count"`
	SkillCallCount      float64 `json:"skill_call_count"`
	ToolAvgDurationMs   float64 `json:"tool_avg_duration_ms"`
	MCPAvgDurationMs    float64 `json:"mcp_avg_duration_ms"`

	// Agent metrics (M-015~016)
	AgentParticipationCount float64 `json:"agent_participation_count"`
	AgentNames              string  `json:"agent_names"`

	// Message metrics (M-017~019)
	TotalMessageCount  float64 `json:"total_message_count"`
	HumanMessageCount  float64 `json:"human_message_count"`
	AgentMessageCount  float64 `json:"agent_message_count"`

	// Engagement metrics (M-020~022)
	HumanParticipationRatio   float64 `json:"human_participation_ratio"`
	HumanAvgIntervalSeconds  float64 `json:"human_avg_interval_seconds"`
	HumanMessageAvgChars      float64 `json:"human_message_avg_chars"`

	// Sentiment metrics (M-023~025)
	HumanNegativeRatio  float64 `json:"human_negative_ratio"`
	HumanAttitudeScore  float64 `json:"human_attitude_score"`
	HumanApprovalRatio  float64 `json:"human_approval_ratio"`

	TaskStatus string `json:"task_status"`
	AnalyzedAt string `json:"analyzed_at"`
}

// AgentParticipation represents an agent's participation in a session
type AgentParticipation struct {
	SessionID     string `json:"session_id"`
	AgentName     string `json:"agent_name"`
	MessageCount  int    `json:"message_count"`
	ToolCallCount int    `json:"tool_call_count"`
}

// ToolCallRecord represents a single tool call record
type ToolCallRecord struct {
	SessionID  string `json:"session_id"`
	Type       string `json:"type"` // TOOL/MCP/SKILL
	Name       string `json:"name"`
	DurationMs int64  `json:"duration_ms"` // 0 if N/A
	Success    bool   `json:"success"`
	CalledAt   string `json:"called_at"`
}

// ReflectionLog represents a reflection event log
type ReflectionLog struct {
	TriggerType      string `json:"trigger_type"`
	TriggerDetail    string `json:"trigger_detail"`
	SessionsAnalyzed int    `json:"sessions_analyzed"`
	ReflectorTokens  int64  `json:"reflector_tokens"`
	DurationMs       int64  `json:"duration_ms"`
	ReportPath       string `json:"report_path"`
	Status           string `json:"status"` // SUCCESS/PARTIAL/FAILED
	TriggeredAt      string `json:"triggered_at"`
}

// Watermark represents the last reflection timestamp for a tool type
type Watermark struct {
	ToolType        string `json:"tool_type"`
	LastReflectedAt string `json:"last_reflected_at"` // ISO 8601
}
