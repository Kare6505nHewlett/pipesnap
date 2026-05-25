// Package retry provides a writer wrapper that retries failed writes
// up to a configurable number of attempts with optional backoff.
package retry

import (
	"errors"
	"fmt"
	"io"
	"time"
)

// ErrMaxAttemptsExceeded is returned when all retry attempts are exhausted.
var ErrMaxAttemptsExceeded = errors.New("retry: max attempts exceeded")

// Config holds retry behaviour parameters.
type Config struct {
	// MaxAttempts is the total number of tries (including the first). Must be >= 1.
	MaxAttempts int
	// Delay is the fixed pause between consecutive attempts. Zero means no pause.
	Delay time.Duration
}

// Retrier wraps an io.Writer and retries writes on error.
type Retrier struct {
	cfg Config
	dst io.Writer
	attempts int
	drops    int
}

// New returns a Retrier that forwards writes to dst, retrying up to
// cfg.MaxAttempts times before returning ErrMaxAttemptsExceeded.
func New(dst io.Writer, cfg Config) (*Retrier, error) {
	if dst == nil {
		return nil, errors.New("retry: dst must not be nil")
	}
	if cfg.MaxAttempts < 1 {
		return nil, fmt.Errorf("retry: MaxAttempts must be >= 1, got %d", cfg.MaxAttempts)
	}
	return &Retrier{cfg: cfg, dst: dst}, nil
}

// Write attempts to write p to the underlying writer, retrying on failure.
func (r *Retrier) Write(p []byte) (int, error) {
	var lastErr error
	for i := 0; i < r.cfg.MaxAttempts; i++ {
		r.attempts++
		n, err := r.dst.Write(p)
		if err == nil {
			return n, nil
		}
		lastErr = err
		if i < r.cfg.MaxAttempts-1 && r.cfg.Delay > 0 {
			time.Sleep(r.cfg.Delay)
		}
	}
	r.drops++
	return 0, fmt.Errorf("%w: %w", ErrMaxAttemptsExceeded, lastErr)
}

// Attempts returns the total number of write attempts made (including retries).
func (r *Retrier) Attempts() int { return r.attempts }

// Drops returns the number of chunks that were ultimately dropped after
// exhausting all retry attempts.
func (r *Retrier) Drops() int { return r.drops }
