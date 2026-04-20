package store

import (
	"fmt"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries    int           // maximum number of retries (default 3)
	InitialDelay  time.Duration // initial backoff delay (default 100ms)
	MaxDelay      time.Duration // max backoff delay (default 1s)
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
	}
}

// WithRetry executes fn with retry logic using exponential backoff
func WithRetry(fn func() error, cfg RetryConfig) error {
	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if attempt < cfg.MaxRetries {
				delay := time.Duration(float64(cfg.InitialDelay) * float64(uint(1)<<attempt))
				if delay > cfg.MaxDelay {
					delay = cfg.MaxDelay
				}
				time.Sleep(delay)
				continue
			}
			return fmt.Errorf("after %d retries: %w", cfg.MaxRetries, lastErr)
		}
		return nil
	}
	return fmt.Errorf("after %d retries: %w", cfg.MaxRetries, lastErr)
}
