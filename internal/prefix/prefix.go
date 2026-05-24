// Package prefix provides a transform that prepends a fixed string to every
// chunk. Chunks must be valid UTF-8; non-UTF-8 data is returned unchanged.
package prefix

import (
	"fmt"
	"io"
)

// Prefixer prepends a fixed byte sequence to every chunk it processes.
type Prefixer struct {
	prefix []byte
}

// New returns a Prefixer that prepends p to every chunk.
// Returns an error if p is empty.
func New(p string) (*Prefixer, error) {
	if p == "" {
		return nil, fmt.Errorf("prefix: prefix string must not be empty")
	}
	return &Prefixer{prefix: []byte(p)}, nil
}

// Apply returns a new byte slice with the prefix prepended to data.
func (pf *Prefixer) Apply(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	out := make([]byte, 0, len(pf.prefix)+len(data))
	out = append(out, pf.prefix...)
	out = append(out, data...)
	return out, nil
}

// NewWriter wraps dst and prepends the configured prefix to every Write call.
func NewWriter(dst io.Writer, p string) (io.WriteCloser, error) {
	pf, err := New(p)
	if err != nil {
		return nil, err
	}
	return &writer{dst: dst, pf: pf}, nil
}

type writer struct {
	dst io.Writer
	pf  *Prefixer
}

func (w *writer) Write(p []byte) (int, error) {
	out, err := w.pf.Apply(p)
	if err != nil {
		return 0, err
	}
	if _, err := w.dst.Write(out); err != nil {
		return 0, err
	}
	// Report the original length so callers do not see a short-write error.
	return len(p), nil
}

func (w *writer) Close() error {
	if c, ok := w.dst.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
