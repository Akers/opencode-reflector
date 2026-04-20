package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/akers/opencode-reflector/internal/model"
)

// GetReportDate returns the date string for the report based on trigger type.
// TIME trigger → previous day, MANUAL/EVENTS trigger → today
func GetReportDate(triggerType model.TriggerType, now time.Time) string {
	if triggerType == model.TriggerTypeTime {
		return now.AddDate(0, 0, -1).Format("2006-01-02")
	}
	return now.Format("2006-01-02")
}

// ReportFilePath returns the full path for a report file with the given date.
func ReportFilePath(basePath, date string) string {
	return filepath.Join(basePath, fmt.Sprintf("dayreport-%s.md", date))
}

// formatNA returns "N/A" if value is -1, otherwise returns the formatted value.
func formatNA(value float64, format string) string {
	if value < 0 {
		return "N/A"
	}
	return fmt.Sprintf(format, value)
}

// formatNAText returns "不可用" if value is -1, otherwise returns the formatted value.
func formatNAText(value float64) string {
	if value < 0 {
		return "不可用"
	}
	return fmt.Sprintf("%.2f%%", value)
}

// formatNAInt returns "N/A" if value is -1, otherwise returns the integer part.
func formatNAInt(value float64) string {
	if value < 0 {
		return "N/A"
	}
	return fmt.Sprintf("%.0f", value)
}

// GenerateReport generates a markdown formatted daily report.
func GenerateReport(date string, sessions []model.SessionMetrics, participations map[string][]model.AgentParticipation, toolCalls map[string][]model.ToolCallRecord, reflectorTokens int64) string {
	var sb strings.Builder

	// Title
	sb.WriteString(fmt.Sprintf("# 日报 %s\n\n", date))

	// Section 1: Task Overview
	sb.WriteString("## 一、任务概览\n\n")
	sb.WriteString("| 序号 | 会话ID | 工具类型 | 任务状态 | 时长(秒) |\n")
	sb.WriteString("| --- | --- | --- | --- | --- |\n")
	for i, s := range sessions {
		sessionID := s.SessionID
		if len(sessionID) > 8 {
			sessionID = sessionID[:8]
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
			i+1, sessionID, s.ToolType, s.TaskStatus, formatNA(s.DurationSeconds, "%.1f")))
	}
	sb.WriteString("\n")

	// Section 2: Token Consumption Statistics
	sb.WriteString("## 二、Token 消耗统计\n\n")
	var totalPrompt, totalCompletion, totalTokens, totalRequests float64
	for _, s := range sessions {
		if s.TotalPromptTokens >= 0 {
			totalPrompt += s.TotalPromptTokens
		}
		if s.TotalCompletionTokens >= 0 {
			totalCompletion += s.TotalCompletionTokens
		}
		if s.TotalTokens >= 0 {
			totalTokens += s.TotalTokens
		}
		if s.ModelRequestCount >= 0 {
			totalRequests += s.ModelRequestCount
		}
	}
	sb.WriteString("| 指标 | 数值 |\n")
	sb.WriteString("| --- | --- |\n")
	sb.WriteString(fmt.Sprintf("| 总 Prompt Tokens | %s |\n", formatNA(totalPrompt, "%.0f")))
	sb.WriteString(fmt.Sprintf("| 总 Completion Tokens | %s |\n", formatNA(totalCompletion, "%.0f")))
	sb.WriteString(fmt.Sprintf("| 总 Tokens | %s |\n", formatNA(totalTokens, "%.0f")))
	sb.WriteString(fmt.Sprintf("| 模型请求次数 | %s |\n", formatNA(totalRequests, "%.0f")))
	sb.WriteString("\n")

	// Section 3: Tool Usage Statistics
	sb.WriteString("## 三、工具使用统计\n\n")
	sb.WriteString("| 工具类型 | 工具名称 | 调用次数 | 成功率 | 平均耗时(ms) |\n")
	sb.WriteString("| --- | --- | --- | --- | --- |\n")

	// Aggregate tool calls by type and name
	toolStats := make(map[string]struct {
		count      int
		success    int
		totalDur   int64
		hasDurData bool
	})
	for _, records := range toolCalls {
		for _, r := range records {
			key := r.Type + ":" + r.Name
			stats := toolStats[key]
			stats.count++
			if r.Success {
				stats.success++
			}
			if r.DurationMs > 0 {
				stats.totalDur += r.DurationMs
				stats.hasDurData = true
			}
			toolStats[key] = stats
		}
	}

	// Write aggregated stats
	toolTypes := []string{"TOOL", "MCP", "SKILL"}
	for _, tt := range toolTypes {
		for key, stats := range toolStats {
			if !strings.HasPrefix(key, tt+":") {
				continue
			}
			name := strings.TrimPrefix(key, tt+":")
			successRate := "N/A"
			if stats.count > 0 {
				successRate = fmt.Sprintf("%.1f%%", float64(stats.success)/float64(stats.count)*100)
			}
			avgDur := "N/A"
			if stats.hasDurData && stats.count > 0 {
				avgDur = fmt.Sprintf("%.2f", float64(stats.totalDur)/float64(stats.count))
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s | %s |\n",
				tt, name, stats.count, successRate, avgDur))
		}
	}
	sb.WriteString("\n")

	// Section 4: Human Participation Analysis
	sb.WriteString("## 四、人类参与分析\n\n")
	var totalParticipationRatio, totalAvgInterval, totalAvgChars float64
	var humanSessionCount int
	for _, s := range sessions {
		if s.HumanParticipationRatio >= 0 {
			totalParticipationRatio += s.HumanParticipationRatio
			humanSessionCount++
		}
		if s.HumanAvgIntervalSeconds >= 0 {
			totalAvgInterval += s.HumanAvgIntervalSeconds
		}
		if s.HumanMessageAvgChars >= 0 {
			totalAvgChars += s.HumanMessageAvgChars
		}
	}
	avgParticipation := "N/A"
	if humanSessionCount > 0 {
		avgParticipation = fmt.Sprintf("%.2f%%", totalParticipationRatio/float64(humanSessionCount))
	}
	avgInterval := "N/A"
	if humanSessionCount > 0 {
		avgInterval = fmt.Sprintf("%.2f秒", totalAvgInterval/float64(humanSessionCount))
	}
	avgChars := "N/A"
	if humanSessionCount > 0 {
		avgChars = fmt.Sprintf("%.2f", totalAvgChars/float64(humanSessionCount))
	}

	sb.WriteString("| 指标 | 数值 |\n")
	sb.WriteString("| --- | --- |\n")
	sb.WriteString(fmt.Sprintf("| 参与率 | %s |\n", avgParticipation))
	sb.WriteString(fmt.Sprintf("| 平均消息间隔 | %s |\n", avgInterval))
	sb.WriteString(fmt.Sprintf("| 平均消息长度 | %s |\n", avgChars))
	sb.WriteString(fmt.Sprintf("| 介入次数 | %s |\n", formatNAInt(totalParticipationRatio)))
	sb.WriteString("\n")

	// Section 5: Sentiment Analysis
	sb.WriteString("## 五、情感分析\n\n")
	var totalNegative, totalAttitude, totalApproval float64
	var sentimentSessionCount int
	for _, s := range sessions {
		if s.HumanNegativeRatio >= 0 {
			totalNegative += s.HumanNegativeRatio
			sentimentSessionCount++
		}
		if s.HumanAttitudeScore >= 0 {
			totalAttitude += s.HumanAttitudeScore
		}
		if s.HumanApprovalRatio >= 0 {
			totalApproval += s.HumanApprovalRatio
		}
	}
	negativeRatio := "不可用"
	if sentimentSessionCount > 0 {
		negativeRatio = fmt.Sprintf("%.2f%%", totalNegative/float64(sentimentSessionCount))
	}
	attitudeScore := "不可用"
	if sentimentSessionCount > 0 {
		attitudeScore = fmt.Sprintf("%.2f", totalAttitude/float64(sentimentSessionCount))
	}
	approvalRatio := "不可用"
	if sentimentSessionCount > 0 {
		approvalRatio = fmt.Sprintf("%.2f%%", totalApproval/float64(sentimentSessionCount))
	}

	sb.WriteString("| 指标 | 数值 |\n")
	sb.WriteString("| --- | --- |\n")
	sb.WriteString(fmt.Sprintf("| 负面占比 | %s |\n", negativeRatio))
	sb.WriteString(fmt.Sprintf("| 态度评分 | %s |\n", attitudeScore))
	sb.WriteString(fmt.Sprintf("| 认可度 | %s |\n", approvalRatio))
	sb.WriteString("\n")

	// Section 6: Reflector Cost
	sb.WriteString("## 六、反思工具自身开销\n\n")
	sb.WriteString("| 指标 | 数值 |\n")
	sb.WriteString("| --- | --- |\n")
	sb.WriteString(fmt.Sprintf("| Reflector Tokens | %d |\n", reflectorTokens))
	sb.WriteString("\n")

	// Section 7: Agent Participation Details
	sb.WriteString("## 七、Agent 参与详情\n\n")
	sb.WriteString("| Agent 名称 | 消息数 | 参与率 |\n")
	sb.WriteString("| --- | --- | --- |\n")

	agentStats := make(map[string]struct {
		messageCount int
		sessionCount int
	})
	totalAgentMessages := 0
	for sessionID, agents := range participations {
		for _, a := range agents {
			stats := agentStats[a.AgentName]
			stats.messageCount += a.MessageCount
			stats.sessionCount++
			agentStats[a.AgentName] = stats
			totalAgentMessages += a.MessageCount
		}
		_ = sessionID // sessionID used for iteration
	}

	for name, stats := range agentStats {
		participationRate := "N/A"
		if stats.sessionCount > 0 && len(sessions) > 0 {
			participationRate = fmt.Sprintf("%.2f%%", float64(stats.sessionCount)/float64(len(sessions))*100)
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %s |\n", name, stats.messageCount, participationRate))
	}
	sb.WriteString("\n")

	return sb.String()
}

// SaveReport saves the report content to a file, creating directories if needed.
// If the file exists, content is appended to the end.
func SaveReport(path string, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Open file with append flag, create if doesn't exist
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Write content
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	return nil
}
