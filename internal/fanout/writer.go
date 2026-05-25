package fanout

import (
	"fmt"
	"io"
)

// Writer wraps a Fanout and satisfies io.WriteCloser. It closes all
// destinations that implement io.Closer when Close is called.
type Writer struct {
	f *Fanout
}

// NewWriter returns a Writer backed by the given Fanout.
func NewWriter(f *Fanout) (*Writer, error) {
	if f == nil {
		return nil, fmt.Errorf("fanout: NewWriter requires a non-nil Fanout")
	}
	return &Writer{f: f}, nil
}

// Write delegates to the underlying Fanout.
func (w *Writer) Write(p []byte) (int, error) {
	return w.f.Write(p)
}

// Close calls Close on every destination that implements io.Closer.
// The first error encountered is returned; remaining closers are still called.
func (w *Writer) Close() error {
	var firstErr error
	for _, d := range w.f.dests {
		c, ok := d.Dst.(io.Closer)
		if !ok {
			continue
		}
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("fanout: close %q: %w", d.Name, err)
		}
	}
	return firstErr
}
