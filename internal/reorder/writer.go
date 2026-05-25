package reorder

import (
	"fmt"
	"io"
)

// Writer is an io.WriteCloser that reorders chunks before forwarding them.
type Writer struct {
	r   *Reorder
	dst io.Writer
}

// Write buffers chunk and emits any ready chunks to the underlying writer.
// It returns the original length of p so callers see no short-write errors
// even when the chunk is buffered rather than immediately forwarded.
func (w *Writer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	cp := make([]byte, len(p))
	copy(cp, p)
	if err := w.r.Add(cp); err != nil {
		return 0, fmt.Errorf("reorder writer: %w", err)
	}
	return len(p), nil
}

// Close flushes all buffered chunks to the underlying writer and, if dst
// implements io.Closer, closes it.
func (w *Writer) Close() error {
	if err := w.r.Flush(); err != nil {
		return err
	}
	if c, ok := w.dst.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Buffered returns the number of chunks still waiting to be emitted.
func (w *Writer) Buffered() int { return w.r.Buffered() }
