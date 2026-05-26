// Package cap provides a chunk writer that enforces a hard cap on the total
// number of chunks or bytes written to a downstream destination. Once the cap
// is reached every subsequent Write is silently dropped and the writer is
// considered exhausted.
package cap

import (
	"errors"
	"io"
)

// ErrCapReached is returned by Write after the cap has been exceeded.
var ErrCapReached = errors.New("cap: limit reached")

// Limiter tracks how many chunks / bytes have been written and decides
// whether the next chunk should be allowed through.
type Limiter struct {
	maxChunks int
	maxBytes  int
	chunks    int
	bytes     int
}

// New creates a Limiter. A value of 0 for either field means "unlimited".
func New(maxChunks, maxBytes int) (*Limiter, error) {
	if maxChunks < 0 {
		return nil, errors.New("cap: maxChunks must be >= 0")
	}
	if maxBytes < 0 {
		return nil, errors.New("cap: maxBytes must be >= 0")
	}
	if maxChunks == 0 && maxBytes == 0 {
		return nil, errors.New("cap: at least one of maxChunks or maxBytes must be > 0")
	}
	return &Limiter{maxChunks: maxChunks, maxBytes: maxBytes}, nil
}

// Allow reports whether the given chunk (of size n bytes) should be forwarded.
// If allowed, the internal counters are updated.
func (l *Limiter) Allow(n int) bool {
	if l.maxChunks > 0 && l.chunks >= l.maxChunks {
		return false
	}
	if l.maxBytes > 0 && l.bytes+n > l.maxBytes {
		return false
	}
	l.chunks++
	l.bytes += n
	return true
}

// Chunks returns the number of chunks forwarded so far.
func (l *Limiter) Chunks() int { return l.chunks }

// Bytes returns the total bytes forwarded so far.
func (l *Limiter) Bytes() int { return l.bytes }

// Exhausted reports whether the limiter will reject all future chunks.
func (l *Limiter) Exhausted() bool {
	if l.maxChunks > 0 && l.chunks >= l.maxChunks {
		return true
	}
	if l.maxBytes > 0 && l.bytes >= l.maxBytes {
		return true
	}
	return false
}

// NewWriter wraps dst, forwarding chunks until the cap is reached.
func NewWriter(dst io.Writer, maxChunks, maxBytes int) (*Writer, error) {
	l, err := New(maxChunks, maxBytes)
	if err != nil {
		return nil, err
	}
	return &Writer{dst: dst, lim: l}, nil
}
