package filter

import (
	"io"
)

// Writer wraps an io.Writer and applies a filter Func to each Write call.
// Chunks that are dropped by the filter are silently discarded.
type Writer struct {
	dst    io.Writer
	filter Func
}

// NewWriter creates a Writer that applies f to every chunk before writing to dst.
func NewWriter(dst io.Writer, f Func) *Writer {
	return &Writer{dst: dst, filter: f}
}

// Write applies the filter to p. If the filter keeps the chunk, it is written
// to the underlying writer. If the filter drops the chunk, Write returns
// len(p), nil to satisfy the io.Writer contract without signaling an error.
func (w *Writer) Write(p []byte) (int, error) {
	chunk := make([]byte, len(p))
	copy(chunk, p)

	transformed, keep := w.filter(chunk)
	if !keep {
		return len(p), nil
	}

	_, err := w.dst.Write(transformed)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
