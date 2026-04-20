package engine

import (
	"regexp"
)

// SentimentResult represents the sentiment analysis result
type SentimentResult struct {
	NegativeRatio  float64 `json:"negative_ratio"`   // 0.0~1.0
	AttitudeScore  float64 `json:"attitude_score"`   // 1~10
	ApprovalRatio  float64 `json:"approval_ratio"`   // 0.0~1.0
}

// regex patterns for sanitization
var (
	apiKeyPattern          = regexp.MustCompile(`sk-[a-zA-Z0-9]{20,}`)
	bearerTokenPattern     = regexp.MustCompile(`Bearer [a-zA-Z0-9\-._~+/]+=*`)
	passwordURLPattern     = regexp.MustCompile(`://[^:]+:[^@]+@`)
	genericPasswordPattern = regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[=:]\s*\S+`)
)

// SanitizeMessages removes sensitive information from messages
func SanitizeMessages(messages []string) []string {
	sanitized := make([]string, len(messages))
	for i, msg := range messages {
		sanitizedMsg := msg

		// Redact API keys (sk- followed by 20+ alphanumeric chars)
		sanitizedMsg = apiKeyPattern.ReplaceAllString(sanitizedMsg, "[REDACTED_API_KEY]")

		// Redact Bearer tokens
		sanitizedMsg = bearerTokenPattern.ReplaceAllString(sanitizedMsg, "Bearer [REDACTED]")

		// Redact password in URL (://user:pass@)
		sanitizedMsg = passwordURLPattern.ReplaceAllString(sanitizedMsg, "://[REDACTED]:[REDACTED]@")

		// Redact generic passwords (password=, passwd=, pwd=)
		sanitizedMsg = genericPasswordPattern.ReplaceAllString(sanitizedMsg, "$1=[REDACTED]")

		sanitized[i] = sanitizedMsg
	}
	return sanitized
}