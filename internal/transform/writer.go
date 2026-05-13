package transform

import "io"

// Writer wraps an io.Writer and applies a transform.Func to every Write call.
// Chunks dropped by the Func are silently discarded; the reported n always
// equals the original length so callers do not see spurious short-write errors.
type Writer struct {
	dst io.Writer
	fn  Func
}

// NewWriter returns a Writer that transforms each chunk with fn before
// forwarding it to dst.
func NewWriter(dst io.Writer, fn Func) *Writer {
	return &Writer{dst: dst, fn: fn}
}

// Write applies the transform function and forwards the result to the
// underlying writer. If the transform drops the chunk, Write returns
// (len(p), nil) without writing anything.
func (w *Writer) Write(p []byte) (int, error) {
	transformed, keep := w.fn(p)
	if !keep {
		return len(p), nil
	}
	if _, err := w.dst.Write(transformed); err != nil {
		return 0, err
	}
	return len(p), nil
}
