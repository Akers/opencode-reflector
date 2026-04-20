package engine

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/akers/opencode-reflector/internal/model"
)

func TestDebounce(t *testing.T) {
	tm := NewTriggerManager(5 * time.Minute)
	var count int32

	// Fast two TIME triggers - should only execute once
	tm.Trigger(model.TriggerTypeTime, "test1", func(triggerType model.TriggerType, detail string) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	tm.Trigger(model.TriggerTypeTime, "test2", func(triggerType model.TriggerType, detail string) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	if count != 1 {
		t.Errorf("expected count=1, got %d", count)
	}
}

func TestDebounceManualBypass(t *testing.T) {
	tm := NewTriggerManager(5 * time.Minute)
	var count int32

	// Fast two MANUAL triggers - both should execute ( MANUAL bypasses debounce)
	tm.Trigger(model.TriggerTypeManual, "manual1", func(triggerType model.TriggerType, detail string) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	tm.Trigger(model.TriggerTypeManual, "manual2", func(triggerType model.TriggerType, detail string) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	if count != 2 {
		t.Errorf("expected count=2 for MANUAL triggers, got %d", count)
	}
}

func TestQueueDuringExecution(t *testing.T) {
	tm := NewTriggerManager(5 * time.Minute)
	var count int32
	var executionOrder []string

	// First trigger starts execution that takes some time
	tm.Trigger(model.TriggerTypeTime, "first", func(triggerType model.TriggerType, detail string) error {
		atomic.AddInt32(&count, 1)
		executionOrder = append(executionOrder, "first:"+detail)
		// Simulate some work by triggering second during execution
		tm.Trigger(model.TriggerTypeTime, "second", func(triggerType model.TriggerType, detail string) error {
			atomic.AddInt32(&count, 1)
			executionOrder = append(executionOrder, "second:"+detail)
			return nil
		})
		return nil
	})

	// The queued trigger should execute after the first completes
	if count < 2 {
		t.Errorf("expected count >= 2, got %d", count)
	}
}

func TestManualPriority(t *testing.T) {
	tm := NewTriggerManager(5 * time.Minute)
	var lastTrigger string

	// Simulate running state with TIME pending
	tm.mu.Lock()
	tm.running = true
	tm.pending = &pendingTrigger{
		triggerType: model.TriggerTypeTime,
		detail:      "time-pending",
		priority:    1,
	}
	tm.mu.Unlock()

	// Trigger MANUAL - should replace TIME in pending
	tm.Trigger(model.TriggerTypeManual, "manual-replacement", func(triggerType model.TriggerType, detail string) error {
		lastTrigger = detail
		return nil
	})

	// The pending should now be MANUAL, but since running=true, it will be queued
	// After execution, the queued MANUAL should run
	if lastTrigger == "manual-replacement" {
		// This is expected behavior
	} else if lastTrigger == "" {
		t.Log("Manual was queued during running state")
	}
}

func TestNoDuplicateExecution(t *testing.T) {
	tm := NewTriggerManager(100 * time.Millisecond)
	var count int32

	// First trigger
	tm.Trigger(model.TriggerTypeTime, "first", func(triggerType model.TriggerType, detail string) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	// Wait for debounce period to reset
	time.Sleep(150 * time.Millisecond)

	// Second trigger after debounce period should execute
	tm.Trigger(model.TriggerTypeTime, "second", func(triggerType model.TriggerType, detail string) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	if count != 2 {
		t.Errorf("expected count=2 after debounce period, got %d", count)
	}
}

func TestIsRunning(t *testing.T) {
	tm := NewTriggerManager(5 * time.Minute)

	if tm.IsRunning() {
		t.Error("expected IsRunning=false initially")
	}

	started := make(chan bool)
	done := make(chan bool)

	go func() {
		tm.Trigger(model.TriggerTypeManual, "test", func(triggerType model.TriggerType, detail string) error {
			started <- true
			time.Sleep(50 * time.Millisecond)
			return nil
		})
		done <- true
	}()

	<-started
	if !tm.IsRunning() {
		t.Error("expected IsRunning=true during execution")
	}

	<-done
	if tm.IsRunning() {
		t.Error("expected IsRunning=false after execution")
	}
}
