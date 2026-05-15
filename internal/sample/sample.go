// Package sample provides a rate-based chunk sampler that forwards only
// every Nth chunk to an underlying io.Writer, useful for reducing snapshot
// volume on high-throughput pipelines.
package sample

import (
	"fmt"
	"io"
)

// Sampler forwards every Nth chunk and discards the rest.
type Sampler struct {
	n       int
	count   int
	dropped int
}

// New creates a Sampler that passes through every nth chunk.
// n must be >= 1; n=1 means pass everything through.
func New(n int) (*Sampler, error) {
	if n < 1 {
		return nil, fmt.Errorf("sample: n must be >= 1, got %d", n)
	}
	return &Sampler{n: n}, nil
}

// Sample returns true if the current chunk should be forwarded.
// It increments the internal counter on every call.
func (s *Sampler) Sample() bool {
	s.count++
	if s.count%s.n == 0 {
		return true
	}
	s.dropped++
	return false
}

// Dropped returns the total number of chunks that were suppressed.
func (s *Sampler) Dropped() int {
	return s.dropped
}

// Seen returns the total number of chunks evaluated so far.
func (s *Sampler) Seen() int {
	return s.count
}

// NewWriter wraps dst so that only every nth chunk is written.
func NewWriter(dst io.Writer, n int) (io.Writer, error) {
	s, err := New(n)
	if err != nil {
		return nil, err
	}
	return &samplerWriter{dst: dst, s: s}, nil
}

type samplerWriter struct {
	dst io.Writer
	s   *Sampler
}

func (w *samplerWriter) Write(p []byte) (int, error) {
	if !w.s.Sample() {
		// Pretend success so callers don't treat a drop as an error.
		return len(p), nil
	}
	return w.dst.Write(p)
}
