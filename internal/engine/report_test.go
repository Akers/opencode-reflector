package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/akers/opencode-reflector/internal/model"
)

func TestGetReportDate(t *testing.T) {
	now := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		trigger    model.TriggerType
		wantDate   string
	}{
		{
			name:       "TIME trigger returns previous day",
			trigger:    model.TriggerTypeTime,
			wantDate:   "2026-04-19",
		},
		{
			name:       "MANUAL trigger returns today",
			trigger:    model.TriggerTypeManual,
			wantDate:   "2026-04-20",
		},
		{
			name:       "EVENTS trigger returns today",
			trigger:    model.TriggerTypeEvents,
			wantDate:   "2026-04-20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetReportDate(tt.trigger, now)
			if got != tt.wantDate {
				t.Errorf("GetReportDate() = %v, want %v", got, tt.wantDate)
			}
		})
	}
}

func TestReportFilename(t *testing.T) {
	date := "2026-04-19"
	got := ReportFilePath("/reports", date)
	want := "/reports/dayreport-2026-04-19.md"

	if got != want {
		t.Errorf("ReportFilePath() = %v, want %v", got, want)
	}
}

func TestGenerateReport(t *testing.T) {
	date := "2026-04-19"
	sessions := []model.SessionMetrics{
		{
			SessionID:               "session-001-abcdef",
			ToolType:                "EDITOR",
			TaskStatus:              "COMPLETED",
			DurationSeconds:         120.5,
			TotalPromptTokens:       1000,
			TotalCompletionTokens:   500,
			TotalTokens:             1500,
			ModelRequestCount:       10,
			ToolCallCount:           5,
			ToolSuccessCount:        4,
			HumanParticipationRatio: 0.6,
			HumanAvgIntervalSeconds: 30.0,
			HumanMessageAvgChars:    50.0,
			HumanNegativeRatio:      0.1,
			HumanAttitudeScore:      4.5,
			HumanApprovalRatio:      0.8,
		},
		{
			SessionID:               "session-002-ghijkl",
			ToolType:                "TERMINAL",
			TaskStatus:              "INTERRUPTED",
			DurationSeconds:         60.0,
			TotalPromptTokens:       800,
			TotalCompletionTokens:   400,
			TotalTokens:             1200,
			ModelRequestCount:       8,
			ToolCallCount:           3,
			ToolSuccessCount:        3,
			HumanParticipationRatio: -1, // N/A
			HumanAvgIntervalSeconds: -1,
			HumanMessageAvgChars:    -1,
			HumanNegativeRatio:      -1,
			HumanAttitudeScore:      -1,
			HumanApprovalRatio:      -1,
		},
	}

	participations := map[string][]model.AgentParticipation{
		"session-001-abcdef": {
			{AgentName: "agent-alpha", MessageCount: 20},
			{AgentName: "agent-beta", MessageCount: 15},
		},
		"session-002-ghijkl": {
			{AgentName: "agent-alpha", MessageCount: 10},
		},
	}

	toolCalls := map[string][]model.ToolCallRecord{
		"session-001-abcdef": {
			{Type: "TOOL", Name: "read_file", DurationMs: 100, Success: true},
			{Type: "TOOL", Name: "write_file", DurationMs: 50, Success: true},
			{Type: "MCP", Name: "git_commit", DurationMs: 200, Success: true},
			{Type: "SKILL", Name: "code_review", DurationMs: 500, Success: false},
			{Type: "TOOL", Name: "no_duration_tool", DurationMs: 0, Success: false},
		},
		"session-002-ghijkl": {
			{Type: "TOOL", Name: "bash_exec", DurationMs: 80, Success: true},
		},
	}

	reflectorTokens := int64(500)

	content := GenerateReport(date, sessions, participations, toolCalls, reflectorTokens)

	// Check title contains date
	if !contains(content, "# 日报 2026-04-19") {
		t.Error("Report should contain title with date")
	}

	// Check all 7 section headers
	sections := []string{
		"## 一、任务概览",
		"## 二、Token 消耗统计",
		"## 三、工具使用统计",
		"## 四、人类参与分析",
		"## 五、情感分析",
		"## 六、反思工具自身开销",
		"## 七、Agent 参与详情",
	}
	for _, section := range sections {
		if !contains(content, section) {
			t.Errorf("Report should contain section: %s", section)
		}
	}

	// Check token values
	if !contains(content, "1800") { // Total tokens: 1500 + 1200
		t.Error("Report should contain total tokens 1800")
	}

	// Check N/A values appear in tool stats (since we have tools with varying success)
	if !contains(content, "N/A") {
		t.Error("Report should contain N/A for unavailable values in tool stats")
	}

	// Check reflector tokens in section 6
	if !contains(content, "500") {
		t.Error("Report should contain reflector tokens value")
	}
}

func TestGenerateReportEmpty(t *testing.T) {
	date := "2026-04-19"
	sessions := []model.SessionMetrics{}
	participations := map[string][]model.AgentParticipation{}
	toolCalls := map[string][]model.ToolCallRecord{}
	reflectorTokens := int64(0)

	content := GenerateReport(date, sessions, participations, toolCalls, reflectorTokens)

	// Should still contain the framework
	if !contains(content, "# 日报 2026-04-19") {
		t.Error("Empty report should still contain title")
	}

	// Should contain all 7 sections
	sections := []string{
		"## 一、任务概览",
		"## 二、Token 消耗统计",
		"## 三、工具使用统计",
		"## 四、人类参与分析",
		"## 五、情感分析",
		"## 六、反思工具自身开销",
		"## 七、Agent 参与详情",
	}
	for _, section := range sections {
		if !contains(content, section) {
			t.Errorf("Empty report should contain section: %s", section)
		}
	}
}

func TestReportReflectorCost(t *testing.T) {
	date := "2026-04-19"
	sessions := []model.SessionMetrics{}
	participations := map[string][]model.AgentParticipation{}
	toolCalls := map[string][]model.ToolCallRecord{}
	reflectorTokens := int64(500)

	content := GenerateReport(date, sessions, participations, toolCalls, reflectorTokens)

	// Check section 6 contains the reflector tokens
	section6Start := "## 六、反思工具自身开销"
	section7Start := "## 七、Agent 参与详情"
	section6Content := content[strings.Index(content, section6Start):strings.Index(content, section7Start)]

	if !contains(section6Content, "500") {
		t.Error("Section 6 should contain reflector tokens value 500")
	}
}

func TestAppendReport(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "dayreport-2026-04-19.md")

	// First write
	content1 := "# 日报 2026-04-19\n\n## 一、任务概览\n\nfirst content\n"
	err := SaveReport(reportPath, content1)
	if err != nil {
		t.Fatalf("First SaveReport failed: %v", err)
	}

	// Second write (append)
	content2 := "second content\n"
	err = SaveReport(reportPath, content2)
	if err != nil {
		t.Fatalf("Second SaveReport failed: %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	combined := string(data)
	if !contains(combined, "first content") {
		t.Error("Combined content should contain first content")
	}
	if !contains(combined, "second content") {
		t.Error("Combined content should contain second content")
	}
}

func TestSaveReportNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "reports", "dayreport-2026-04-19.md")

	content := "# 日报 2026-04-19\n\ntest content\n"
	err := SaveReport(reportPath, content)
	if err != nil {
		t.Fatalf("SaveReport failed: %v", err)
	}

	// Verify file exists
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != content {
		t.Errorf("File content mismatch: got %v, want %v", string(data), content)
	}
}

func TestSaveReportCreatesDir(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "nested", "dirs", "dayreport-2026-04-19.md")

	content := "# 日报 2026-04-19\n\ntest content\n"
	err := SaveReport(reportPath, content)
	if err != nil {
		t.Fatalf("SaveReport failed: %v", err)
	}

	// Verify file exists
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != content {
		t.Errorf("File content mismatch: got %v, want %v", string(data), content)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
