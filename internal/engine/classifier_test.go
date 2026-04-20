package engine

import (
	"testing"

	"github.com/akers/opencode-reflector/internal/model"
)

// TestL1ArchiveDetected tests that /opsx:archive in last human message returns COMPLETED
func TestL1ArchiveDetected(t *testing.T) {
	session := model.CanonicalSession{
		Messages: []model.CanonicalMessage{
			{Role: model.MessageRoleHuman, Content: "Hello"},
			{Role: model.MessageRoleAgent, Content: "Hi there"},
			{Role: model.MessageRoleHuman, Content: "Please /opsx:archive this session"},
		},
	}

	status, ok := checkL1(session.Messages)
	if status != model.TaskStatusCompleted {
		t.Errorf("expected COMPLETED, got %v", status)
	}
	if !ok {
		t.Error("expected ok to be true")
	}
}

// TestL1NoArchive tests that missing /opsx:archive returns false
func TestL1NoArchive(t *testing.T) {
	session := model.CanonicalSession{
		Messages: []model.CanonicalMessage{
			{Role: model.MessageRoleHuman, Content: "Hello"},
			{Role: model.MessageRoleAgent, Content: "Hi there"},
			{Role: model.MessageRoleHuman, Content: "Just a regular message"},
		},
	}

	status, ok := checkL1(session.Messages)
	if status != model.TaskStatusUncertain {
		t.Errorf("expected UNCERTAIN, got %v", status)
	}
	if ok {
		t.Error("expected ok to be false")
	}
}

// TestL3ChineseKeywords tests Chinese completion keywords
func TestL3ChineseKeywords(t *testing.T) {
	testCases := []struct {
		content  string
		expected model.TaskStatus
		ok       bool
	}{
		{"任务已完成", model.TaskStatusCompleted, true},
		{"完成工作", model.TaskStatusCompleted, true},
		{"工作完成", model.TaskStatusCompleted, true},
		{"任务完成", model.TaskStatusCompleted, true},
	}

	for _, tc := range testCases {
		status, ok := checkL3Chinese(tc.content)
		if status != tc.expected {
			t.Errorf("content=%q: expected %v, got %v", tc.content, tc.expected, status)
		}
		if ok != tc.ok {
			t.Errorf("content=%q: expected ok=%v, got %v", tc.content, tc.ok, ok)
		}
	}
}

// TestL3EnglishKeywords tests English completion keywords (case insensitive)
func TestL3EnglishKeywords(t *testing.T) {
	testCases := []struct {
		content  string
		expected model.TaskStatus
		ok       bool
	}{
		{"Task completed", model.TaskStatusCompleted, true},
		{"TASK COMPLETED", model.TaskStatusCompleted, true},
		{"All done", model.TaskStatusCompleted, true},
		{"Finished", model.TaskStatusCompleted, true},
		{"Task complete", model.TaskStatusCompleted, true},
		{"Completed", model.TaskStatusCompleted, true},
	}

	for _, tc := range testCases {
		status, ok := checkL3English(tc.content)
		if status != tc.expected {
			t.Errorf("content=%q: expected %v, got %v", tc.content, tc.expected, status)
		}
		if ok != tc.ok {
			t.Errorf("content=%q: expected ok=%v, got %v", tc.content, tc.ok, ok)
		}
	}
}

// TestL3NegativeContext tests that negative context prevents COMPLETED status
func TestL3NegativeContext(t *testing.T) {
	testCases := []struct {
		content  string
		expected model.TaskStatus
		ok       bool
	}{
		{"无法完成", model.TaskStatusUncertain, false},
		{"未能完成任务", model.TaskStatusUncertain, false},
		{"不能完成这个任务", model.TaskStatusUncertain, false},
		{"没有完成工作", model.TaskStatusUncertain, false},
	}

	for _, tc := range testCases {
		status, ok := checkL3Chinese(tc.content)
		if status != tc.expected {
			t.Errorf("content=%q: expected %v, got %v", tc.content, tc.expected, status)
		}
		if ok != tc.ok {
			t.Errorf("content=%q: expected ok=%v, got %v", tc.content, tc.ok, ok)
		}
	}
}

// TestClassifyTaskCascade tests the full cascading classification
func TestClassifyTaskCascade(t *testing.T) {
	// L1 hit case
	sessionL1 := model.CanonicalSession{
		Messages: []model.CanonicalMessage{
			{Role: model.MessageRoleHuman, Content: "Hello"},
			{Role: model.MessageRoleHuman, Content: "/opsx:archive"},
		},
	}
	if status := ClassifyTask(sessionL1); status != model.TaskStatusCompleted {
		t.Errorf("L1 hit: expected COMPLETED, got %v", status)
	}

	// L3 hit case
	sessionL3 := model.CanonicalSession{
		Messages: []model.CanonicalMessage{
			{Role: model.MessageRoleHuman, Content: "Hello"},
			{Role: model.MessageRoleAgent, Content: "Task completed successfully"},
		},
	}
	if status := ClassifyTask(sessionL3); status != model.TaskStatusCompleted {
		t.Errorf("L3 hit: expected COMPLETED, got %v", status)
	}

	// No hit case - should return UNCERTAIN
	sessionNone := model.CanonicalSession{
		Messages: []model.CanonicalMessage{
			{Role: model.MessageRoleHuman, Content: "Hello"},
			{Role: model.MessageRoleAgent, Content: "I am thinking..."},
		},
	}
	if status := ClassifyTask(sessionNone); status != model.TaskStatusUncertain {
		t.Errorf("No hit: expected UNCERTAIN, got %v", status)
	}
}

// TestClassifyTaskL2FromMetadata tests L2 classification from metadata with high confidence
func TestClassifyTaskL2FromMetadata(t *testing.T) {
	session := model.CanonicalSession{
		Metadata: map[string]interface{}{
			"task_classification": map[string]interface{}{
				"status":     "interrupted",
				"confidence": 0.85,
			},
		},
		Messages: []model.CanonicalMessage{
			{Role: model.MessageRoleHuman, Content: "Hello"},
			{Role: model.MessageRoleAgent, Content: "Working on it"},
		},
	}

	if status := ClassifyTask(session); status != model.TaskStatusInterrupted {
		t.Errorf("L2 hit: expected INTERRUPTED, got %v", status)
	}
}

// TestClassifyTaskL2LowConfidence tests that low confidence falls back to L3
func TestClassifyTaskL2LowConfidence(t *testing.T) {
	session := model.CanonicalSession{
		Metadata: map[string]interface{}{
			"task_classification": map[string]interface{}{
				"status":     "interrupted",
				"confidence": 0.5,
			},
		},
		Messages: []model.CanonicalMessage{
			{Role: model.MessageRoleHuman, Content: "Hello"},
			{Role: model.MessageRoleAgent, Content: "Task completed successfully"},
		},
	}

	// Should fall back to L3 which matches "Task completed successfully"
	if status := ClassifyTask(session); status != model.TaskStatusCompleted {
		t.Errorf("L2 low confidence: expected COMPLETED (L3 fallback), got %v", status)
	}
}

// TestClassifyTaskL2InvalidStatus tests that empty status falls back to L3
func TestClassifyTaskL2InvalidStatus(t *testing.T) {
	session := model.CanonicalSession{
		Metadata: map[string]interface{}{
			"task_classification": map[string]interface{}{
				"status":     "",
				"confidence": 0.9,
			},
		},
		Messages: []model.CanonicalMessage{
			{Role: model.MessageRoleHuman, Content: "Hello"},
			{Role: model.MessageRoleAgent, Content: "Task completed successfully"},
		},
	}

	// Should fall back to L3 which matches "Task completed successfully"
	if status := ClassifyTask(session); status != model.TaskStatusCompleted {
		t.Errorf("L2 invalid status: expected COMPLETED (L3 fallback), got %v", status)
	}
}

// TestAbandonedStatus tests that last human message returns ABANDONED
func TestAbandonedStatus(t *testing.T) {
	session := model.CanonicalSession{
		Messages: []model.CanonicalMessage{
			{Role: model.MessageRoleHuman, Content: "Hello"},
			{Role: model.MessageRoleAgent, Content: "Hi there"},
			{Role: model.MessageRoleHuman, Content: "I need to leave"},
		},
	}

	if status := DetermineTaskStatus(session); status != model.TaskStatusAbandoned {
		t.Errorf("expected ABANDONED, got %v", status)
	}
}

// TestNotAbandoned tests that last agent message uses ClassifyTask
func TestNotAbandoned(t *testing.T) {
	// Last message is agent, and L1 would hit with archive
	session := model.CanonicalSession{
		Messages: []model.CanonicalMessage{
			{Role: model.MessageRoleHuman, Content: "Hello"},
			{Role: model.MessageRoleAgent, Content: "Task completed"},
		},
	}

	// Should use ClassifyTask, not ABANDONED
	if status := DetermineTaskStatus(session); status == model.TaskStatusAbandoned {
		t.Errorf("expected NOT ABANDONED, got %v", status)
	}
}
