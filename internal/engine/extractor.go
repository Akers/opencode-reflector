package engine

import (
	"time"

	"github.com/akers/opencode-reflector/internal/model"
)

// ExtractTimeMetrics extracts time-related metrics from a session
func ExtractTimeMetrics(session *model.CanonicalSession) (startedAt, endedAt string, durationSeconds, humanThinkTimeSeconds, agentThinkTimeSeconds float64) {
	if session == nil || len(session.Messages) == 0 {
		return "", "", 0, 0, 0
	}

	messages := session.Messages
	startedAt = messages[0].Timestamp
	endedAt = messages[len(messages)-1].Timestamp

	startTime, err1 := time.Parse(time.RFC3339, startedAt)
	endTime, err2 := time.Parse(time.RFC3339, endedAt)
	if err1 == nil && err2 == nil {
		durationSeconds = endTime.Sub(startTime).Seconds()
	}

	// Calculate human think time (adjacent human messages)
	var humanIntervals []float64
	var lastHumanTime *time.Time
	for _, msg := range messages {
		if msg.Role == model.MessageRoleHuman {
			msgTime, err := time.Parse(time.RFC3339, msg.Timestamp)
			if err != nil {
				continue
			}
			if lastHumanTime != nil {
				humanIntervals = append(humanIntervals, msgTime.Sub(*lastHumanTime).Seconds())
			}
			lastHumanTime = &msgTime
		}
	}
	for _, interval := range humanIntervals {
		humanThinkTimeSeconds += interval
	}

	// Calculate agent think time (adjacent agent messages)
	var agentIntervals []float64
	var lastAgentTime *time.Time
	for _, msg := range messages {
		if msg.Role == model.MessageRoleAgent {
			msgTime, err := time.Parse(time.RFC3339, msg.Timestamp)
			if err != nil {
				continue
			}
			if lastAgentTime != nil {
				agentIntervals = append(agentIntervals, msgTime.Sub(*lastAgentTime).Seconds())
			}
			lastAgentTime = &msgTime
		}
	}
	for _, interval := range agentIntervals {
		agentThinkTimeSeconds += interval
	}

	return startedAt, endedAt, durationSeconds, humanThinkTimeSeconds, agentThinkTimeSeconds
}

// ExtractTokenMetrics extracts token-related metrics from a session
func ExtractTokenMetrics(session *model.CanonicalSession) (promptTokens, completionTokens, totalTokens, modelRequestCount float64) {
	if session == nil || len(session.Messages) == 0 {
		return -1, -1, -1, -1
	}

	var totalPrompt, totalCompletion int
	var messagesWithTokens int

	for _, msg := range session.Messages {
		if msg.PromptTokens != nil {
			totalPrompt += *msg.PromptTokens
			messagesWithTokens++
		}
		if msg.CompletionTokens != nil {
			totalCompletion += *msg.CompletionTokens
		}
	}

	// If no token data found, return -1 for all
	if messagesWithTokens == 0 {
		return -1, -1, -1, -1
	}

	promptTokens = float64(totalPrompt)
	completionTokens = float64(totalCompletion)
	totalTokens = float64(totalPrompt + totalCompletion)
	modelRequestCount = float64(messagesWithTokens)

	return promptTokens, completionTokens, totalTokens, modelRequestCount
}

// ExtractToolMetrics extracts tool call metrics from a session
func ExtractToolMetrics(session *model.CanonicalSession) (toolCallCount, toolSuccessCount, mcpCallCount, skillCallCount, toolAvgDurationMs, mcpAvgDurationMs float64, records []model.ToolCallRecord) {
	if session == nil {
		return 0, 0, 0, 0, -1, -1, nil
	}

	var toolDurations []int64
	var mcpDurations []int64

	for _, tc := range session.ToolCalls {
		records = append(records, model.ToolCallRecord{
			SessionID:  session.ID,
			Type:       string(tc.Type),
			Name:       tc.Name,
			DurationMs: 0,
			Success:    tc.Success,
			CalledAt:   tc.CalledAt,
		})

		if tc.DurationMs != nil {
			records[len(records)-1].DurationMs = *tc.DurationMs
		}

		switch tc.Type {
		case model.ToolCallTypeTool:
			toolCallCount++
			if tc.Success {
				toolSuccessCount++
			}
			if tc.DurationMs != nil {
				toolDurations = append(toolDurations, *tc.DurationMs)
			}
		case model.ToolCallTypeMCP:
			mcpCallCount++
			if tc.DurationMs != nil {
				mcpDurations = append(mcpDurations, *tc.DurationMs)
			}
		case model.ToolCallTypeSkill:
			skillCallCount++
		}
	}

	// Calculate averages
	if len(toolDurations) > 0 {
		var sum int64
		for _, d := range toolDurations {
			sum += d
		}
		toolAvgDurationMs = float64(sum) / float64(len(toolDurations))
	} else {
		toolAvgDurationMs = -1
	}

	if len(mcpDurations) > 0 {
		var sum int64
		for _, d := range mcpDurations {
			sum += d
		}
		mcpAvgDurationMs = float64(sum) / float64(len(mcpDurations))
	} else {
		mcpAvgDurationMs = -1
	}

	return toolCallCount, toolSuccessCount, mcpCallCount, skillCallCount, toolAvgDurationMs, mcpAvgDurationMs, records
}

// ExtractAgentMetrics extracts agent participation metrics from a session
func ExtractAgentMetrics(session *model.CanonicalSession) (participationCount float64, participations []model.AgentParticipation, agentNames string) {
	if session == nil || len(session.Messages) == 0 {
		return 0, nil, ""
	}

	agentMessageCount := make(map[string]int)
	for _, msg := range session.Messages {
		if msg.Role == model.MessageRoleAgent && msg.AgentName != nil {
			agentMessageCount[*msg.AgentName]++
		}
	}

	if len(agentMessageCount) == 0 {
		return 0, nil, ""
	}

	var names []string
	for name, count := range agentMessageCount {
		participations = append(participations, model.AgentParticipation{
			SessionID:    session.ID,
			AgentName:    name,
			MessageCount: count,
		})
		names = append(names, name)
	}

	participationCount = float64(len(participations))

	// Build comma-separated agent names
	for i, name := range names {
		if i > 0 {
			agentNames += ","
		}
		agentNames += name
	}

	return participationCount, participations, agentNames
}

// ExtractMessageMetrics extracts message count metrics from a session
func ExtractMessageMetrics(session *model.CanonicalSession) (total, agentCount, humanCount float64) {
	if session == nil {
		return 0, 0, 0
	}

	var systemCount int
	for _, msg := range session.Messages {
		switch msg.Role {
		case model.MessageRoleAgent:
			agentCount++
		case model.MessageRoleHuman:
			humanCount++
		case model.MessageRoleSystem:
			systemCount++
		}
	}

	total = humanCount + agentCount

	return total, agentCount, humanCount
}

// ExtractHumanParticipation extracts human participation metrics from a session
func ExtractHumanParticipation(session *model.CanonicalSession) (ratio float64, avgIntervalSeconds float64, avgChars float64) {
	if session == nil || len(session.Messages) == 0 {
		return 0, -1, -1
	}

	var humanMessages []model.CanonicalMessage
	for _, msg := range session.Messages {
		if msg.Role == model.MessageRoleHuman {
			humanMessages = append(humanMessages, msg)
		}
	}

	if len(humanMessages) == 0 {
		return 0, -1, -1
	}

	// Calculate ratio
	var agentCount int
	for _, msg := range session.Messages {
		if msg.Role == model.MessageRoleAgent {
			agentCount++
		}
	}

	total := float64(len(humanMessages)) + float64(agentCount)
	if total > 0 {
		ratio = float64(len(humanMessages)) / total
	}

	// Calculate average interval between human messages
	if len(humanMessages) > 1 {
		var totalInterval float64
		for i := 1; i < len(humanMessages); i++ {
			t1, err1 := time.Parse(time.RFC3339, humanMessages[i-1].Timestamp)
			t2, err2 := time.Parse(time.RFC3339, humanMessages[i].Timestamp)
			if err1 == nil && err2 == nil {
				totalInterval += t2.Sub(t1).Seconds()
			}
		}
		avgIntervalSeconds = totalInterval / float64(len(humanMessages)-1)
	} else {
		avgIntervalSeconds = -1
	}

	// Calculate average character count
	var totalChars int
	for _, msg := range humanMessages {
		totalChars += len(msg.Content)
	}
	avgChars = float64(totalChars) / float64(len(humanMessages))

	return ratio, avgIntervalSeconds, avgChars
}

// ExtractAll extracts all metrics from a session
func ExtractAll(session *model.CanonicalSession) *model.SessionMetrics {
	metrics := &model.SessionMetrics{
		SessionID: session.ID,
		ToolType:  string(session.ToolType),
		Title:     session.Title,
		TaskStatus: "", // Phase 6
	}

	if session.Title != nil {
		metrics.Title = session.Title
	} else {
		metrics.Title = nil
	}

	// Time metrics
	startedAt, endedAt, durationSeconds, humanThinkTimeSeconds, agentThinkTimeSeconds := ExtractTimeMetrics(session)
	metrics.StartedAt = startedAt
	metrics.EndedAt = endedAt
	metrics.DurationSeconds = durationSeconds
	metrics.HumanThinkTimeSeconds = humanThinkTimeSeconds
	metrics.AgentThinkTimeSeconds = agentThinkTimeSeconds

	// Token metrics
	promptTokens, completionTokens, totalTokens, modelRequestCount := ExtractTokenMetrics(session)
	metrics.TotalPromptTokens = promptTokens
	metrics.TotalCompletionTokens = completionTokens
	metrics.TotalTokens = totalTokens
	metrics.ModelRequestCount = modelRequestCount

	// Tool metrics
	toolCallCount, toolSuccessCount, mcpCallCount, skillCallCount, toolAvgDurationMs, mcpAvgDurationMs, _ := ExtractToolMetrics(session)
	metrics.ToolCallCount = toolCallCount
	metrics.ToolSuccessCount = toolSuccessCount
	metrics.MCPCallCount = mcpCallCount
	metrics.SkillCallCount = skillCallCount
	metrics.ToolAvgDurationMs = toolAvgDurationMs
	metrics.MCPAvgDurationMs = mcpAvgDurationMs

	// Agent metrics
	agentParticipationCount, _, agentNames := ExtractAgentMetrics(session)
	metrics.AgentParticipationCount = agentParticipationCount
	metrics.AgentNames = agentNames

	// Message metrics
	totalMessageCount, agentMessageCount, humanMessageCount := ExtractMessageMetrics(session)
	metrics.TotalMessageCount = totalMessageCount
	metrics.AgentMessageCount = agentMessageCount
	metrics.HumanMessageCount = humanMessageCount

	// Human participation metrics
	humanRatio, humanAvgInterval, humanAvgChars := ExtractHumanParticipation(session)
	metrics.HumanParticipationRatio = humanRatio
	metrics.HumanAvgIntervalSeconds = humanAvgInterval
	metrics.HumanMessageAvgChars = humanAvgChars

	// Active time (same as duration for now, can be enhanced in Phase 6)
	metrics.ActiveTimeSeconds = durationSeconds

	// Sentiment metrics (Phase 6) - leave as -1
	metrics.HumanNegativeRatio = -1
	metrics.HumanAttitudeScore = -1
	metrics.HumanApprovalRatio = -1

	metrics.AnalyzedAt = time.Now().Format(time.RFC3339)

	return metrics
}
