package retry

import (
	"errors"
	"io"
)

// Writer is an io.WriteCloser that wraps a Retrier and optionally closes
// the underlying destination when Close is called.
type Writer struct {
	r   *Retrier
	dst io.Writer
}

// NewWriter constructs a Writer that retries failed writes to dst according
// to cfg. If dst also implements io.Closer, Close will propagate to it.
func NewWriter(dst io.Writer, cfg Config) (*Writer, error) {
	r, err := New(dst, cfg)
	if err != nil {
		return nil, err
	}
	return &Writer{r: r, dst: dst}, nil
}

// Write forwards p through the underlying Retrier.
func (w *Writer) Write(p []byte) (int, error) {
	return w.r.Write(p)
}

// Attempts returns the total write attempts recorded by the Retrier.
func (w *Writer) Attempts() int { return w.r.Attempts() }

// Drops returns the number of chunks dropped after exhausting all retries.
func (w *Writer) Drops() int { return w.r.Drops() }

// Close closes the underlying writer if it implements io.Closer.
func (w *Writer) Close() error {
	if c, ok := w.dst.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Unwrap returns the underlying io.Writer for use with errors.As / errors.Is.
func (w *Writer) Unwrap() error { return errors.New("retry.Writer") }
