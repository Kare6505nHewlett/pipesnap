// Package skip provides a filter that discards the first N chunks
// from a stream, passing all subsequent chunks through unchanged.
package skip

import (
	"fmt"
	"io"
)

// Skipper discards the first N chunks written to it, then forwards
// all remaining chunks to the underlying destination.
type Skipper struct {
	dst       io.Writer
	skipCount int
	seen      int
	dropped   int
}

// New returns a new Skipper that will discard the first n chunks.
// n must be greater than zero.
func New(dst io.Writer, n int) (*Skipper, error) {
	if dst == nil {
		return nil, fmt.Errorf("skip: dst must not be nil")
	}
	if n <= 0 {
		return nil, fmt.Errorf("skip: n must be greater than zero, got %d", n)
	}
	return &Skipper{
		dst:       dst,
		skipCount: n,
	}, nil
}

// Write implements io.Writer. The first n chunks are silently dropped;
// subsequent chunks are forwarded to the destination unchanged.
func (s *Skipper) Write(p []byte) (int, error) {
	s.seen++
	if s.seen <= s.skipCount {
		s.dropped++
		return len(p), nil
	}
	return s.dst.Write(p)
}

// Seen returns the total number of chunks that have been presented.
func (s *Skipper) Seen() int { return s.seen }

// Dropped returns the number of chunks that were skipped.
func (s *Skipper) Dropped() int { return s.dropped }

// Passed returns the number of chunks forwarded to dst.
func (s *Skipper) Passed() int { return s.seen - s.dropped }

// NewWriter wraps New and panics on error, for use in initialisation
// contexts where the arguments are known to be valid.
func NewWriter(dst io.Writer, n int) *Skipper {
	s, err := New(dst, n)
	if err != nil {
		panic(err)
	}
	return s
}
