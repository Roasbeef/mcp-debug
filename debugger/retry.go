package debugger

import (
	"context"
	"fmt"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultRetryConfig provides sensible defaults for retry operations
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:  5,
	InitialDelay: 10 * time.Millisecond,
	MaxDelay:     500 * time.Millisecond,
	Multiplier:   2.0,
}

// RetryWithBackoff executes a function with exponential backoff retry logic
func RetryWithBackoff(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error
	delay := config.InitialDelay
	
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		err := operation()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Don't wait after the last attempt
		if attempt == config.MaxAttempts {
			break
		}
		
		// Wait with exponential backoff
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		
		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}
	
	return fmt.Errorf("operation failed after %d attempts, last error: %w", config.MaxAttempts, lastErr)
}