// Package throttle provides a token-bucket-style chunk throttler that limits
// the number of chunks emitted per time window, independent of byte size.
package throttle

import (
	"errors"
	"io"
	"sync"
	"time"
)

// Throttle limits how many chunks may pass through per window.
type Throttle struct {
	mu       sync.Mutex
	max      int           // max chunks per window
	current  int           // chunks emitted in current window
	window   time.Duration // length of each window
	lastReset time.Time
}

// New creates a Throttle that allows at most maxPerWindow chunks per window.
// window must be positive and maxPerWindow must be >= 1.
func New(maxPerWindow int, window time.Duration) (*Throttle, error) {
	if maxPerWindow < 1 {
		return nil, errors.New("throttle: maxPerWindow must be >= 1")
	}
	if window <= 0 {
		return nil, errors.New("throttle: window must be positive")
	}
	return &Throttle{
		max:       maxPerWindow,
		window:    window,
		lastReset: time.Now(),
	}, nil
}

// Allow reports whether a chunk may pass. If the window has elapsed it resets
// the counter first. Returns false when the budget for the current window is
// exhausted.
func (t *Throttle) Allow() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if now.Sub(t.lastReset) >= t.window {
		t.current = 0
		t.lastReset = now
	}
	if t.current >= t.max {
		return false
	}
	t.current++
	return true
}

// Remaining returns the number of chunks still allowed in the current window.
func (t *Throttle) Remaining() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	if time.Since(t.lastReset) >= t.window {
		return t.max
	}
	r := t.max - t.current
	if r < 0 {
		return 0
	}
	return r
}

// NewWriter wraps dst, forwarding only chunks that the throttle allows.
// Dropped chunks are silently discarded; the original length is still returned
// so callers do not treat a drop as an error.
func NewWriter(dst io.Writer, th *Throttle) io.Writer {
	return &writer{dst: dst, th: th}
}

type writer struct {
	dst io.Writer
	th  *Throttle
}

func (w *writer) Write(p []byte) (int, error) {
	if !w.th.Allow() {
		return len(p), nil
	}
	return w.dst.Write(p)
}
