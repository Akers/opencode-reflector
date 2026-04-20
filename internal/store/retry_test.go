package store

import (
	"fmt"
	"testing"
	"time"
)

func TestRetrySuccessFirstTry(t *testing.T) {
	calls := 0
	err := WithRetry(func() error {
		calls++
		return nil
	}, RetryConfig{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetrySuccessAfterFailures(t *testing.T) {
	calls := 0
	err := WithRetry(func() error {
		calls++
		if calls < 3 {
			return fmt.Errorf("mock error")
		}
		return nil
	}, RetryConfig{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetryAllFailed(t *testing.T) {
	calls := 0
	mockErr := fmt.Errorf("mock error")
	err := WithRetry(func() error {
		calls++
		return mockErr
	}, RetryConfig{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should be called MaxRetries + 1 times (initial + retries)
	if calls != 4 {
		t.Errorf("expected 4 calls, got %d", calls)
	}
	// Verify error message contains retry count
	if !contains(err.Error(), "after 3 retries") {
		t.Errorf("error message should contain retry count: %v", err)
	}
}

func TestRetryDefaultConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	if cfg.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", cfg.MaxRetries)
	}
	if cfg.InitialDelay != 100*time.Millisecond {
		t.Errorf("expected InitialDelay 100ms, got %v", cfg.InitialDelay)
	}
	if cfg.MaxDelay != 1*time.Second {
		t.Errorf("expected MaxDelay 1s, got %v", cfg.MaxDelay)
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
