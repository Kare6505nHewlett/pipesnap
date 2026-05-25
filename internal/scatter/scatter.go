// Package scatter distributes chunks across multiple destination writers
// using a round-robin or hash-based strategy.
package scatter

import (
	"errors"
	"fmt"
	"hash/fnv"
	"io"
)

// Strategy controls how chunks are distributed across destinations.
type Strategy int

const (
	// RoundRobin sends successive chunks to successive destinations in order.
	RoundRobin Strategy = iota
	// HashBased routes each chunk to a destination determined by its content hash.
	HashBased
)

// Scatter distributes chunks across a fixed set of io.Writers.
type Scatter struct {
	dests    []io.Writer
	strategy Strategy
	cursor   int
}

// New creates a Scatter that fans chunks out across dests using the given strategy.
// At least one destination must be provided and none may be nil.
func New(strategy Strategy, dests ...io.Writer) (*Scatter, error) {
	if len(dests) == 0 {
		return nil, errors.New("scatter: at least one destination required")
	}
	for i, d := range dests {
		if d == nil {
			return nil, fmt.Errorf("scatter: destination %d is nil", i)
		}
	}
	return &Scatter{dests: dests, strategy: strategy}, nil
}

// Write sends p to exactly one destination writer according to the configured
// strategy. It returns len(p), nil on success.
func (s *Scatter) Write(p []byte) (int, error) {
	idx := s.pick(p)
	n, err := s.dests[idx].Write(p)
	if err != nil {
		return n, fmt.Errorf("scatter: dest %d: %w", idx, err)
	}
	if n != len(p) {
		return n, fmt.Errorf("scatter: dest %d: short write", idx)
	}
	return n, nil
}

// Len returns the number of destination writers.
func (s *Scatter) Len() int { return len(s.dests) }

func (s *Scatter) pick(p []byte) int {
	switch s.strategy {
	case HashBased:
		h := fnv.New32a()
		_, _ = h.Write(p)
		return int(h.Sum32()) % len(s.dests)
	default: // RoundRobin
		idx := s.cursor
		s.cursor = (s.cursor + 1) % len(s.dests)
		return idx
	}
}
