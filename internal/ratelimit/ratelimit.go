// Package ratelimit provides a token-bucket style rate limiter for
// controlling the throughput of replayed snapshot chunks.
package ratelimit

import (
	"time"
)

// Limiter controls the rate at which chunks are emitted during replay.
type Limiter struct {
	bytesPerSec int64
	minDelay    time.Duration
}

// Config holds configuration for the rate limiter.
type Config struct {
	// BytesPerSec caps throughput; 0 means unlimited.
	BytesPerSec int64
	// MinDelay enforces a minimum pause between chunks regardless of size.
	MinDelay time.Duration
}

// New returns a Limiter configured with the given Config.
// If both fields are zero the limiter is a no-op.
func New(cfg Config) *Limiter {
	return &Limiter{
		bytesPerSec: cfg.BytesPerSec,
		minDelay:    cfg.MinDelay,
	}
}

// Wait blocks for the appropriate duration before a chunk of size n bytes
// is allowed to proceed.
func (l *Limiter) Wait(n int) {
	delay := l.minDelay

	if l.bytesPerSec > 0 && n > 0 {
		throttle := time.Duration(float64(time.Second) * float64(n) / float64(l.bytesPerSec))
		if throttle > delay {
			delay = throttle
		}
	}

	if delay > 0 {
		time.Sleep(delay)
	}
}

// IsNoOp returns true when the limiter will never introduce a delay.
func (l *Limiter) IsNoOp() bool {
	return l.bytesPerSec == 0 && l.minDelay == 0
}
