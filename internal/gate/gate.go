// Package gate provides a conditional pass-through writer that opens or
// closes the stream based on a user-supplied predicate. While the gate is
// closed every incoming chunk is silently dropped; once opened chunks flow
// through to the destination unchanged.
package gate

import (
	"errors"
	"io"
	"sync"
)

// Predicate decides whether a chunk should pass through the gate.
type Predicate func(data []byte) bool

// Gate controls whether chunks are forwarded to the destination writer.
type Gate struct {
	mu     sync.Mutex
	open   bool
	pred   Predicate
	dst    io.Writer
	seen   int
	passed int
}

// New creates a Gate that starts closed. The supplied predicate is called on
// every chunk; the gate opens permanently on the first chunk for which the
// predicate returns true.
//
// dst must not be nil. pred must not be nil.
func New(dst io.Writer, pred Predicate) (*Gate, error) {
	if dst == nil {
		return nil, errors.New("gate: dst must not be nil")
	}
	if pred == nil {
		return nil, errors.New("gate: pred must not be nil")
	}
	return &Gate{dst: dst, pred: pred}, nil
}

// Open forces the gate into the open state regardless of the predicate.
func (g *Gate) Open() {
	g.mu.Lock()
	g.open = true
	g.mu.Unlock()
}

// Close forces the gate into the closed state.
func (g *Gate) Close() {
	g.mu.Lock()
	g.open = false
	g.mu.Unlock()
}

// IsOpen reports whether the gate is currently open.
func (g *Gate) IsOpen() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.open
}

// Write evaluates the predicate and forwards p to the destination when the
// gate is open. It always returns len(p), nil so the caller's pipeline is
// never interrupted by a closed gate.
func (g *Gate) Write(p []byte) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.seen++
	if !g.open {
		if g.pred(p) {
			g.open = true
		} else {
			return len(p), nil
		}
	}

	g.passed++
	_, err := g.dst.Write(p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Seen returns the total number of chunks presented to the gate.
func (g *Gate) Seen() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.seen
}

// Passed returns the number of chunks that were forwarded to the destination.
func (g *Gate) Passed() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.passed
}

// Stats returns a snapshot of the seen and passed chunk counts in a single
// atomic read, avoiding two separate lock acquisitions when both values are
// needed together.
func (g *Gate) Stats() (seen, passed int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.seen, g.passed
}
