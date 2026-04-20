package store

import (
	"context"
	"time"

	"github.com/akers/opencode-reflector/internal/model"
)

// Store defines the persistence interface for reflector data
type Store interface {
	// Session operations
	InsertSession(ctx context.Context, metrics *model.SessionMetrics) error
	GetSessionsByDateRange(ctx context.Context, start, end time.Time) ([]model.SessionMetrics, error)
	GetSessionsByStatus(ctx context.Context, status model.TaskStatus) ([]model.SessionMetrics, error)

	// Related data operations
	InsertAgentParticipations(ctx context.Context, sessionID string, participations []model.AgentParticipation) error
	InsertToolCalls(ctx context.Context, sessionID string, calls []model.ToolCallRecord) error

	// Reflection log operations
	InsertReflectionLog(ctx context.Context, log *model.ReflectionLog) error

	// Watermark operations
	GetWatermark(ctx context.Context) (time.Time, error)
	UpdateWatermark(ctx context.Context, t time.Time) error

	// Cleanup
	Cleanup(ctx context.Context, days int) error

	// Lifecycle
	Close() error
}
