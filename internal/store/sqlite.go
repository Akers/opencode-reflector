package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
	"github.com/akers/opencode-reflector/internal/model"
)

// SQLiteStore implements Store interface using SQLite
type SQLiteStore struct {
	db *sql.DB
}

// NewStore creates a new SQLite store with the given database path
func NewStore(ctx context.Context, dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.createTables(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return store, nil
}

// createTables creates all required tables if they don't exist
func (s *SQLiteStore) createTables(ctx context.Context) error {
	ddlStatements := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id              TEXT PRIMARY KEY,
			agent_tool      TEXT NOT NULL,
			title           TEXT,
			start_time      DATETIME NOT NULL,
			end_time        DATETIME NOT NULL,
			status          TEXT NOT NULL,
			duration_seconds REAL DEFAULT 0,
			active_time_seconds REAL DEFAULT 0,
			human_think_time_seconds REAL DEFAULT 0,
			agent_think_time_seconds REAL DEFAULT 0,
			prompt_tokens   INTEGER DEFAULT 0,
			completion_tokens INTEGER DEFAULT 0,
			total_tokens    INTEGER,
			llm_request_count INTEGER DEFAULT 0,
			total_messages  INTEGER DEFAULT 0,
			agent_messages  INTEGER DEFAULT 0,
			human_messages  INTEGER DEFAULT 0,
			human_participation_rate REAL DEFAULT 0,
			human_avg_interval_seconds REAL,
			human_avg_chars REAL DEFAULT 0,
			tool_call_count REAL DEFAULT 0,
			tool_success_count REAL DEFAULT 0,
			mcp_call_count  REAL DEFAULT 0,
			skill_call_count REAL DEFAULT 0,
			tool_avg_duration_ms REAL,
			mcp_avg_duration_ms REAL,
			negative_emotion_ratio REAL DEFAULT 0,
			attitude_score  REAL,
			human_approval_ratio REAL,
			agent_names     TEXT,
			agent_count     REAL DEFAULT 0,
			reflector_tokens INTEGER DEFAULT 0,
			created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS agent_participations (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id      TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			agent_name      TEXT NOT NULL,
			message_count   INTEGER NOT NULL,
			participation_rate REAL NOT NULL,
			created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tool_calls (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id      TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			call_type       TEXT NOT NULL,
			tool_name       TEXT NOT NULL,
			duration_ms     INTEGER,
			success         BOOLEAN DEFAULT TRUE,
			called_at       DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS reflection_logs (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			trigger_type    TEXT NOT NULL,
			trigger_detail  TEXT,
			sessions_count  INTEGER DEFAULT 0,
			tokens_consumed INTEGER DEFAULT 0,
			status          TEXT NOT NULL,
			error_message   TEXT,
			started_at      DATETIME NOT NULL,
			completed_at    DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS watermarks (
			id              INTEGER PRIMARY KEY CHECK (id = 1),
			last_reflection_time DATETIME NOT NULL,
			updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_end_time ON sessions(end_time)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_tool_calls_session ON tool_calls(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tool_calls_type_name ON tool_calls(call_type, tool_name)`,
	}

	for _, ddl := range ddlStatements {
		if _, err := s.db.ExecContext(ctx, ddl); err != nil {
			return fmt.Errorf("failed to execute DDL: %w", err)
		}
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// InsertSession inserts a session metrics record
func (s *SQLiteStore) InsertSession(ctx context.Context, metrics *model.SessionMetrics) error {
	// Convert float64 to *int64 for token fields (nil if -1)
	var promptTokens, completionTokens, llmRequestCount, totalTokens sql.NullInt64
	if metrics.TotalPromptTokens >= 0 {
		promptTokens = sql.NullInt64{Int64: int64(metrics.TotalPromptTokens), Valid: true}
	}
	if metrics.TotalCompletionTokens >= 0 {
		completionTokens = sql.NullInt64{Int64: int64(metrics.TotalCompletionTokens), Valid: true}
	}
	if metrics.ModelRequestCount >= 0 {
		llmRequestCount = sql.NullInt64{Int64: int64(metrics.ModelRequestCount), Valid: true}
	}
	if metrics.TotalTokens >= 0 {
		totalTokens = sql.NullInt64{Int64: int64(metrics.TotalTokens), Valid: true}
	}

	// Handle optional title
	var title sql.NullString
	if metrics.Title != nil {
		title = sql.NullString{String: *metrics.Title, Valid: true}
	}

	// Convert float64 to nullable for avg duration fields (-1 = NULL)
	var toolAvgDurationMs, mcpAvgDurationMs, humanAvgIntervalSeconds, humanApprovalRatio sql.NullFloat64
	if metrics.ToolAvgDurationMs >= 0 {
		toolAvgDurationMs = sql.NullFloat64{Float64: metrics.ToolAvgDurationMs, Valid: true}
	}
	if metrics.MCPAvgDurationMs >= 0 {
		mcpAvgDurationMs = sql.NullFloat64{Float64: metrics.MCPAvgDurationMs, Valid: true}
	}
	if metrics.HumanAvgIntervalSeconds >= 0 {
		humanAvgIntervalSeconds = sql.NullFloat64{Float64: metrics.HumanAvgIntervalSeconds, Valid: true}
	}
	if metrics.HumanApprovalRatio >= 0 {
		humanApprovalRatio = sql.NullFloat64{Float64: metrics.HumanApprovalRatio, Valid: true}
	}

	// attitude_score (HumanAttitudeScore)
	var attitudeScore sql.NullFloat64
	if metrics.HumanAttitudeScore >= 0 {
		attitudeScore = sql.NullFloat64{Float64: metrics.HumanAttitudeScore, Valid: true}
	}

	query := `INSERT INTO sessions (
		id, agent_tool, title, start_time, end_time, status,
		duration_seconds, active_time_seconds, human_think_time_seconds, agent_think_time_seconds,
		prompt_tokens, completion_tokens, total_tokens, llm_request_count,
		total_messages, agent_messages, human_messages,
		human_participation_rate, human_avg_interval_seconds, human_avg_chars,
		tool_call_count, tool_success_count, mcp_call_count, skill_call_count,
		tool_avg_duration_ms, mcp_avg_duration_ms,
		negative_emotion_ratio, attitude_score, human_approval_ratio,
		agent_names, agent_count, reflector_tokens
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query,
		metrics.SessionID,
		metrics.ToolType,
		title,
		metrics.StartedAt,
		metrics.EndedAt,
		metrics.TaskStatus,
		metrics.DurationSeconds,
		metrics.ActiveTimeSeconds,
		metrics.HumanThinkTimeSeconds,
		metrics.AgentThinkTimeSeconds,
		promptTokens,
		completionTokens,
		totalTokens,
		llmRequestCount,
		int(metrics.TotalMessageCount),
		int(metrics.AgentMessageCount),
		int(metrics.HumanMessageCount),
		metrics.HumanParticipationRatio,
		humanAvgIntervalSeconds,
		metrics.HumanMessageAvgChars,
		metrics.ToolCallCount,
		metrics.ToolSuccessCount,
		metrics.MCPCallCount,
		metrics.SkillCallCount,
		toolAvgDurationMs,
		mcpAvgDurationMs,
		metrics.HumanNegativeRatio,
		attitudeScore,
		humanApprovalRatio,
		metrics.AgentNames,
		metrics.AgentParticipationCount,
		0, // reflector_tokens not in metrics
	)
	if err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}

	return nil
}

// GetSessionsByDateRange retrieves sessions within a date range
func (s *SQLiteStore) GetSessionsByDateRange(ctx context.Context, start, end time.Time) ([]model.SessionMetrics, error) {
	query := `SELECT
		id, agent_tool, title, start_time, end_time, status,
		duration_seconds, active_time_seconds, human_think_time_seconds, agent_think_time_seconds,
		prompt_tokens, completion_tokens, total_tokens, llm_request_count,
		total_messages, agent_messages, human_messages,
		human_participation_rate, human_avg_interval_seconds, human_avg_chars,
		tool_call_count, tool_success_count, mcp_call_count, skill_call_count,
		tool_avg_duration_ms, mcp_avg_duration_ms,
		negative_emotion_ratio, attitude_score, human_approval_ratio,
		agent_names, agent_count, reflector_tokens
	FROM sessions
	WHERE start_time >= ? AND start_time <= ?`

	rows, err := s.db.QueryContext(ctx, query, start.Format(time.RFC3339), end.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	return scanSessions(rows)
}

// GetSessionsByStatus retrieves sessions with a specific status
func (s *SQLiteStore) GetSessionsByStatus(ctx context.Context, status model.TaskStatus) ([]model.SessionMetrics, error) {
	query := `SELECT
		id, agent_tool, title, start_time, end_time, status,
		duration_seconds, active_time_seconds, human_think_time_seconds, agent_think_time_seconds,
		prompt_tokens, completion_tokens, total_tokens, llm_request_count,
		total_messages, agent_messages, human_messages,
		human_participation_rate, human_avg_interval_seconds, human_avg_chars,
		tool_call_count, tool_success_count, mcp_call_count, skill_call_count,
		tool_avg_duration_ms, mcp_avg_duration_ms,
		negative_emotion_ratio, attitude_score, human_approval_ratio,
		agent_names, agent_count, reflector_tokens
	FROM sessions
	WHERE status = ?`

	rows, err := s.db.QueryContext(ctx, query, string(status))
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	return scanSessions(rows)
}

// scanSessions scans rows into SessionMetrics slice
func scanSessions(rows *sql.Rows) ([]model.SessionMetrics, error) {
	var sessions []model.SessionMetrics
	for rows.Next() {
		var s model.SessionMetrics
		var title sql.NullString
		var promptTokens, completionTokens, llmRequestCount, totalTokens sql.NullInt64
		var agentCount sql.NullFloat64
		var humanParticipationRate, humanAvgIntervalSeconds, humanAvgChars sql.NullFloat64
		var toolCallCount, toolSuccessCount, mcpCallCount, skillCallCount sql.NullFloat64
		var toolAvgDurationMs, mcpAvgDurationMs sql.NullFloat64
		var negativeEmotionRatio, attitudeScore, humanApprovalRatio sql.NullFloat64
		var agentNames sql.NullString
		var reflectorTokens sql.NullInt64

		err := rows.Scan(
			&s.SessionID,
			&s.ToolType,
			&title,
			&s.StartedAt,
			&s.EndedAt,
			&s.TaskStatus,
			&s.DurationSeconds,
			&s.ActiveTimeSeconds,
			&s.HumanThinkTimeSeconds,
			&s.AgentThinkTimeSeconds,
			&promptTokens,
			&completionTokens,
			&totalTokens,
			&llmRequestCount,
			&s.TotalMessageCount,
			&s.AgentMessageCount,
			&s.HumanMessageCount,
			&humanParticipationRate,
			&humanAvgIntervalSeconds,
			&humanAvgChars,
			&toolCallCount,
			&toolSuccessCount,
			&mcpCallCount,
			&skillCallCount,
			&toolAvgDurationMs,
			&mcpAvgDurationMs,
			&negativeEmotionRatio,
			&attitudeScore,
			&humanApprovalRatio,
			&agentNames,
			&agentCount,
			&reflectorTokens,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		if title.Valid {
			s.Title = &title.String
		}
		if promptTokens.Valid {
			s.TotalPromptTokens = float64(promptTokens.Int64)
		}
		if completionTokens.Valid {
			s.TotalCompletionTokens = float64(completionTokens.Int64)
		}
		if totalTokens.Valid {
			s.TotalTokens = float64(totalTokens.Int64)
		}
		if llmRequestCount.Valid {
			s.ModelRequestCount = float64(llmRequestCount.Int64)
		}
		if humanParticipationRate.Valid {
			s.HumanParticipationRatio = humanParticipationRate.Float64
		}
		if humanAvgIntervalSeconds.Valid {
			s.HumanAvgIntervalSeconds = humanAvgIntervalSeconds.Float64
		}
		if humanAvgChars.Valid {
			s.HumanMessageAvgChars = humanAvgChars.Float64
		}
		if toolCallCount.Valid {
			s.ToolCallCount = toolCallCount.Float64
		}
		if toolSuccessCount.Valid {
			s.ToolSuccessCount = toolSuccessCount.Float64
		}
		if mcpCallCount.Valid {
			s.MCPCallCount = mcpCallCount.Float64
		}
		if skillCallCount.Valid {
			s.SkillCallCount = skillCallCount.Float64
		}
		if toolAvgDurationMs.Valid {
			s.ToolAvgDurationMs = toolAvgDurationMs.Float64
		}
		if mcpAvgDurationMs.Valid {
			s.MCPAvgDurationMs = mcpAvgDurationMs.Float64
		}
		if negativeEmotionRatio.Valid {
			s.HumanNegativeRatio = negativeEmotionRatio.Float64
		}
		if attitudeScore.Valid {
			s.HumanAttitudeScore = attitudeScore.Float64
		}
		if humanApprovalRatio.Valid {
			s.HumanApprovalRatio = humanApprovalRatio.Float64
		}
		if agentNames.Valid {
			s.AgentNames = agentNames.String
		}
		if agentCount.Valid {
			s.AgentParticipationCount = agentCount.Float64
		}

		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// InsertAgentParticipations inserts agent participation records for a session
func (s *SQLiteStore) InsertAgentParticipations(ctx context.Context, sessionID string, participations []model.AgentParticipation) error {
	query := `INSERT INTO agent_participations (session_id, agent_name, message_count, participation_rate) VALUES (?, ?, ?, ?)`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, p := range participations {
		// Map ToolCallCount to participation_rate as per specification
		participationRate := float64(p.ToolCallCount)
		_, err := stmt.ExecContext(ctx, sessionID, p.AgentName, p.MessageCount, participationRate)
		if err != nil {
			return fmt.Errorf("failed to insert agent participation: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// InsertToolCalls inserts tool call records for a session
func (s *SQLiteStore) InsertToolCalls(ctx context.Context, sessionID string, calls []model.ToolCallRecord) error {
	query := `INSERT INTO tool_calls (session_id, call_type, tool_name, duration_ms, success, called_at) VALUES (?, ?, ?, ?, ?, ?)`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, c := range calls {
		var durationMs sql.NullInt64
		if c.DurationMs > 0 {
			durationMs = sql.NullInt64{Int64: c.DurationMs, Valid: true}
		}

		_, err := stmt.ExecContext(ctx, sessionID, c.Type, c.Name, durationMs, c.Success, c.CalledAt)
		if err != nil {
			return fmt.Errorf("failed to insert tool call: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// InsertReflectionLog inserts a reflection log record
func (s *SQLiteStore) InsertReflectionLog(ctx context.Context, log *model.ReflectionLog) error {
	query := `INSERT INTO reflection_logs (trigger_type, trigger_detail, sessions_count, tokens_consumed, status, error_message, started_at, completed_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	var errorMsg sql.NullString
	if log.ReportPath != "" {
		errorMsg = sql.NullString{String: log.ReportPath, Valid: true}
	}

	_, err := s.db.ExecContext(ctx, query,
		log.TriggerType,
		log.TriggerDetail,
		log.SessionsAnalyzed,
		log.ReflectorTokens,
		log.Status,
		errorMsg,
		log.TriggeredAt,
		time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert reflection log: %w", err)
	}

	return nil
}

// GetWatermark retrieves the last reflection timestamp
func (s *SQLiteStore) GetWatermark(ctx context.Context) (time.Time, error) {
	query := `SELECT last_reflection_time FROM watermarks WHERE id = 1`

	var watermarkTime time.Time
	err := s.db.QueryRowContext(ctx, query).Scan(&watermarkTime)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get watermark: %w", err)
	}

	return watermarkTime, nil
}

// UpdateWatermark updates or inserts the watermark timestamp
func (s *SQLiteStore) UpdateWatermark(ctx context.Context, t time.Time) error {
	query := `INSERT OR REPLACE INTO watermarks (id, last_reflection_time, updated_at) VALUES (1, ?, ?)`

	_, err := s.db.ExecContext(ctx, query, t.Format(time.RFC3339), time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to update watermark: %w", err)
	}

	return nil
}

// Cleanup removes sessions older than the specified number of days
func (s *SQLiteStore) Cleanup(ctx context.Context, days int) error {
	// Delete sessions older than the specified days
	// Foreign key constraints will cascade delete related records
	query := `DELETE FROM sessions WHERE end_time < datetime('now', ?)`

	_, err := s.db.ExecContext(ctx, query, fmt.Sprintf("-%d days", days))
	if err != nil {
		return fmt.Errorf("failed to cleanup sessions: %w", err)
	}

	return nil
}

// Ensure Store interface is implemented
var _ Store = (*SQLiteStore)(nil)

// Helper function to convert string to nullable
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// Helper function to convert float64 to nullable int64
func toNullInt64(f float64) sql.NullInt64 {
	if f < 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(f), Valid: true}
}

// Helper to build IN clause placeholder string
func buildInPlaceholders(n int) string {
	if n <= 0 {
		return ""
	}
	return "(" + strings.Repeat("?,", n-1) + "?)"
}
