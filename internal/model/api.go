package model

// ReflectRequest represents a reflection analysis request
type ReflectRequest struct {
	TriggerType   string           `json:"trigger_type"`
	TriggerDetail string           `json:"trigger_detail"`
	Sessions      []CanonicalSession `json:"sessions"`
	Config        ReflectConfig    `json:"config"`
}

// ReflectConfig represents configuration for reflection analysis
type ReflectConfig struct {
	SentimentEnabled bool   `json:"sentiment_enabled"`
	SentimentSource  string `json:"sentiment_source"` // "llm" | "builtin" | "na"
}

// ReflectResponse represents the response from a reflection analysis
type ReflectResponse struct {
	Status           string `json:"status"` // SUCCESS/PARTIAL/FAILED
	SessionsAnalyzed int    `json:"sessions_analyzed"`
	ReportPath       string `json:"report_path"`
	TokensConsumed   int64  `json:"tokens_consumed"`
	DurationMs       int64  `json:"duration_ms"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string `json:"status"` // "ok"
	Version string `json:"version"`
}

// StatsResponse represents statistics about reflection analysis
type StatsResponse struct {
	TotalReflections      int64 `json:"total_reflections"`
	TotalTokensConsumed   int64 `json:"total_tokens_consumed"`
	TotalSessionsAnalyzed int64 `json:"total_sessions_analyzed"`
}

// CleanupRequest represents a request to cleanup old data
type CleanupRequest struct {
	Days int `json:"days"`
}
