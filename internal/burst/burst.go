// Package burst provides a sliding-window rate limiter that allows short
// bursts of chunks up to a configured capacity before throttling writes.
package burst

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Limiter tracks a token-bucket style burst allowance.
type Limiter struct {
	mu       sync.Mutex
	cap      int           // maximum tokens in bucket
	tokens   int           // current available tokens
	refill   int           // tokens added per interval
	interval time.Duration // refill interval
	last     time.Time
}

// New creates a Limiter with the given burst capacity and refill rate.
// cap is the maximum number of chunks allowed in a burst.
// refill tokens are added every interval until cap is reached.
func New(cap, refill int, interval time.Duration) (*Limiter, error) {
	if cap <= 0 {
		return nil, fmt.Errorf("burst: cap must be > 0, got %d", cap)
	}
	if refill <= 0 {
		return nil, fmt.Errorf("burst: refill must be > 0, got %d", refill)
	}
	if interval <= 0 {
		return nil, fmt.Errorf("burst: interval must be > 0")
	}
	return &Limiter{
		cap:      cap,
		tokens:   cap,
		refill:   refill,
		interval: interval,
		last:     time.Now(),
	}, nil
}

// Allow blocks until a token is available, then consumes one.
func (l *Limiter) Allow() {
	for {
		l.mu.Lock()
		l.addTokens()
		if l.tokens > 0 {
			l.tokens--
			l.mu.Unlock()
			return
		}
		wait := l.interval
		l.mu.Unlock()
		time.Sleep(wait / 2)
	}
}

// addTokens refills tokens based on elapsed time. Must be called with mu held.
func (l *Limiter) addTokens() {
	now := time.Now()
	elapsed := now.Sub(l.last)
	if elapsed >= l.interval {
		periods := int(elapsed / l.interval)
		l.tokens += periods * l.refill
		if l.tokens > l.cap {
			l.tokens = l.cap
		}
		l.last = l.last.Add(time.Duration(periods) * l.interval)
	}
}

// NewWriter wraps dst so that each Write call is gated by the Limiter.
func NewWriter(dst io.Writer, l *Limiter) io.Writer {
	return &writer{dst: dst, limiter: l}
}

type writer struct {
	dst     io.Writer
	limiter *Limiter
}

func (w *writer) Write(p []byte) (int, error) {
	w.limiter.Allow()
	return w.dst.Write(p)
}
