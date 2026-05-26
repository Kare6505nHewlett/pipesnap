package gate

import (
	"errors"
	"io"
)

var (
	errNilGate = errors.New("gate: gate must not be nil")
	errNilDst  = errors.New("gate: dst must not be nil")
)

// Writer wraps a Gate and implements io.WriteCloser so it can be composed
// with other pipeline stages that expect a closer.
type Writer struct {
	gate *Gate
	dst  io.Closer
}

// NewWriter creates a Writer that forwards chunks through g and closes dst
// when Close is called.
func NewWriter(g *Gate, dst io.WriteCloser) (*Writer, error) {
	if g == nil {
		return nil, errNilGate
	}
	if dst == nil {
		return nil, errNilDst
	}
	return &Writer{gate: g, dst: dst}, nil
}

// Write passes p through the gate.
func (w *Writer) Write(p []byte) (int, error) {
	return w.gate.Write(p)
}

// Close closes the underlying destination.
func (w *Writer) Close() error {
	return w.dst.Close()
}

// Gate returns the underlying Gate so callers can inspect or control state.
func (w *Writer) Gate() *Gate {
	return w.gate
}
