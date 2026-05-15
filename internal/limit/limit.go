// Package limit provides a chunk-count and byte-size limiter for snapshot
// readers. It wraps an io.Reader and returns io.EOF once the configured
// threshold is reached, allowing callers to consume only a leading slice of a
// snapshot without reading the entire file.
package limit

import (
	"errors"
	"io"
)

// ErrLimitReached is returned when the configured limit has been exceeded.
var ErrLimitReached = errors.New("limit: threshold reached")

// Limiter holds the current consumption state.
type Limiter struct {
	maxChunks int
	maxBytes  int64
	chunks    int
	bytes     int64
}

// New creates a Limiter. Pass 0 to disable a particular threshold.
func New(maxChunks int, maxBytes int64) *Limiter {
	return &Limiter{
		maxChunks: maxChunks,
		maxBytes:  maxBytes,
	}
}

// Allow reports whether the next chunk of size n bytes may be consumed.
// It updates internal counters when the chunk is accepted.
func (l *Limiter) Allow(n int) bool {
	if l.maxChunks > 0 && l.chunks >= l.maxChunks {
		return false
	}
	if l.maxBytes > 0 && l.bytes+int64(n) > l.maxBytes {
		return false
	}
	l.chunks++
	l.bytes += int64(n)
	return true
}

// Chunks returns the number of chunks accepted so far.
func (l *Limiter) Chunks() int { return l.chunks }

// Bytes returns the total bytes accepted so far.
func (l *Limiter) Bytes() int64 { return l.bytes }

// WrapReader wraps r so that Read calls return io.EOF once the limiter
// refuses a chunk. Each call to Read is treated as one chunk.
func WrapReader(r io.Reader, l *Limiter) io.Reader {
	return &limitedReader{r: r, l: l}
}

type limitedReader struct {
	r io.Reader
	l *Limiter
	done bool
}

func (lr *limitedReader) Read(p []byte) (int, error) {
	if lr.done {
		return 0, io.EOF
	}
	n, err := lr.r.Read(p)
	if n > 0 && !lr.l.Allow(n) {
		lr.done = true
		return 0, io.EOF
	}
	if err != nil {
		lr.done = true
	}
	return n, err
}
