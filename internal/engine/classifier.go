package engine

import (
	"strings"

	"github.com/akers/opencode-reflector/internal/model"
)

// checkL1 checks if the last human message contains /opsx:archive command
func checkL1(messages []model.CanonicalMessage) (model.TaskStatus, bool) {
	// Find the last human message
	var lastHumanMsg *model.CanonicalMessage
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == model.MessageRoleHuman {
			lastHumanMsg = &messages[i]
			break
		}
	}

	if lastHumanMsg == nil {
		return model.TaskStatusUncertain, false
	}

	if strings.Contains(lastHumanMsg.Content, "/opsx:archive") {
		return model.TaskStatusCompleted, true
	}

	return model.TaskStatusUncertain, false
}

// checkL3Chinese checks Chinese completion keywords
func checkL3Chinese(content string) (model.TaskStatus, bool) {
	completionKeywords := []string{"任务已完成", "完成工作", "工作完成", "任务完成"}
	negativeKeywords := []string{"无法完成", "未能完成", "不能完成", "没有完成"}

	lowerContent := strings.ToLower(content)

	hasCompletion := false
	hasNegative := false

	for _, kw := range completionKeywords {
		if strings.Contains(lowerContent, strings.ToLower(kw)) {
			hasCompletion = true
			break
		}
	}

	for _, kw := range negativeKeywords {
		if strings.Contains(lowerContent, strings.ToLower(kw)) {
			hasNegative = true
			break
		}
	}

	if hasCompletion && !hasNegative {
		return model.TaskStatusCompleted, true
	}

	return model.TaskStatusUncertain, false
}

// checkL3English checks English completion keywords
func checkL3English(content string) (model.TaskStatus, bool) {
	completionKeywords := []string{"task completed", "all done", "finished", "task complete", "completed"}
	negativeKeywords := []string{"did not complete", "cannot complete", "unable to complete", "not completed", "failed to"}

	lowerContent := strings.ToLower(content)

	hasCompletion := false
	hasNegative := false

	for _, kw := range completionKeywords {
		if strings.Contains(lowerContent, strings.ToLower(kw)) {
			hasCompletion = true
			break
		}
	}

	for _, kw := range negativeKeywords {
		if strings.Contains(lowerContent, strings.ToLower(kw)) {
			hasNegative = true
			break
		}
	}

	if hasCompletion && !hasNegative {
		return model.TaskStatusCompleted, true
	}

	return model.TaskStatusUncertain, false
}

// checkL3 checks the last agent message for completion indicators
func checkL3(messages []model.CanonicalMessage) (model.TaskStatus, bool) {
	// Find the last agent message
	var lastAgentMsg *model.CanonicalMessage
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == model.MessageRoleAgent {
			lastAgentMsg = &messages[i]
			break
		}
	}

	if lastAgentMsg == nil {
		return model.TaskStatusUncertain, false
	}

	// Check English first, then Chinese
	if status, ok := checkL3English(lastAgentMsg.Content); ok {
		return status, true
	}

	if status, ok := checkL3Chinese(lastAgentMsg.Content); ok {
		return status, true
	}

	return model.TaskStatusUncertain, false
}

// ClassifyTask performs cascading L1 -> L2(metadata) -> L3 classification
func ClassifyTask(session model.CanonicalSession) model.TaskStatus {
	messages := session.Messages

	// L1 check
	if status, ok := checkL1(messages); ok {
		return status
	}

	// L2 check (from metadata, injected by Adapter)
	if session.Metadata != nil {
		if tc := model.GetMetadataMap(session.Metadata, "task_classification"); tc != nil {
			status := strings.ToUpper(model.GetMetadataString(tc, "status", ""))
			confidence := model.GetMetadataFloat(tc, "confidence", 0)
			if status != "" && confidence >= 0.7 {
				return model.TaskStatus(status)
			}
		}
	}

	// L3 check (fallback)
	if status, ok := checkL3(messages); ok {
		return status
	}

	return model.TaskStatusUncertain
}

// DetermineTaskStatus determines the final task status
func DetermineTaskStatus(session model.CanonicalSession) model.TaskStatus {
	messages := session.Messages

	// If the last message is from human, the user abandoned the conversation
	if len(messages) > 0 && messages[len(messages)-1].Role == model.MessageRoleHuman {
		return model.TaskStatusAbandoned
	}

	// Otherwise, use cascading classification
	return ClassifyTask(session)
}