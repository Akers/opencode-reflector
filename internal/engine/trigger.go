package engine

import (
	"sync"
	"time"

	"github.com/akers/opencode-reflector/internal/model"
)

// TriggerManager manages trigger execution with debouncing and queuing
type TriggerManager struct {
	mu          sync.Mutex
	lastTrigger time.Time
	debounceMin time.Duration
	running     bool
	pending     *pendingTrigger
}

type pendingTrigger struct {
	triggerType model.TriggerType
	detail      string
	priority    int
}

// NewTriggerManager creates a new TriggerManager with the specified debounce duration
func NewTriggerManager(debounceMin time.Duration) *TriggerManager {
	if debounceMin <= 0 {
		debounceMin = 5 * time.Minute
	}
	return &TriggerManager{
		debounceMin: debounceMin,
	}
}

// Trigger handles a trigger event with debouncing and queuing
func (tm *TriggerManager) Trigger(triggerType model.TriggerType, detail string, execute func(triggerType model.TriggerType, detail string) error) error {
	// Fast path: try lock without blocking to avoid deadlock if execute() calls Trigger()
	if !tm.mu.TryLock() {
		// Another trigger is executing, queue this one if higher priority
		tm.mu.Lock()
		if tm.running {
			newPriority := getPriority(triggerType)
			if tm.pending == nil || newPriority > tm.pending.priority {
				tm.pending = &pendingTrigger{
					triggerType: triggerType,
					detail:      detail,
					priority:    newPriority,
				}
			}
		}
		tm.mu.Unlock()
		return nil
	}

	// Check debounce for non-MANUAL triggers
	if triggerType != model.TriggerTypeManual {
		if time.Since(tm.lastTrigger) < tm.debounceMin {
			tm.mu.Unlock()
			return nil // Ignore due to debounce
		}
	}

	// If already running, queue the trigger instead of concurrent execution
	if tm.running {
		// If there's already a pending trigger and new one has higher priority, replace it
		newPriority := getPriority(triggerType)
		if tm.pending == nil || newPriority > tm.pending.priority {
			tm.pending = &pendingTrigger{
				triggerType: triggerType,
				detail:      detail,
				priority:    newPriority,
			}
		}
		tm.mu.Unlock()
		return nil // Ignore, will be handled after current execution
	}

	// Mark as running and execute (execute runs outside lock to avoid deadlock)
	tm.running = true
	tm.mu.Unlock()

	err := execute(triggerType, detail)

	// Process pending triggers after execution completes
	tm.mu.Lock()
	if tm.pending != nil {
		pending := tm.pending
		tm.pending = nil
		tm.running = true
		tm.mu.Unlock()

		err2 := execute(pending.triggerType, pending.detail)
		if err == nil {
			err = err2
		}

		tm.mu.Lock()
	}

	// Update last trigger time
	tm.lastTrigger = time.Now()
	tm.running = false
	tm.mu.Unlock()

	return err
}

// IsRunning returns whether a trigger is currently being executed
func (tm *TriggerManager) IsRunning() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.running
}

// getPriority returns the priority for a trigger type
func getPriority(triggerType model.TriggerType) int {
	switch triggerType {
	case model.TriggerTypeManual:
		return 10
	case model.TriggerTypeEvents:
		return 5
	case model.TriggerTypeTime:
		return 1
	default:
		return 0
	}
}
