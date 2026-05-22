package aggregate

import (
	"fmt"
	"io"
)

// Writer is an io.WriteCloser that feeds each Write call into an
// Aggregator and flushes any remaining buffered data on Close.
type Writer struct {
	agg *Aggregator
	dst io.Writer
}

// NewWriter returns a Writer that aggregates every size chunks into a
// single write to dst.
func NewWriter(dst io.Writer, size int) (*Writer, error) {
	if dst == nil {
		return nil, fmt.Errorf("aggregate: dst must not be nil")
	}
	agg, err := New(size, func(p []byte) error {
		_, werr := dst.Write(p)
		return werr
	})
	if err != nil {
		return nil, err
	}
	return &Writer{agg: agg, dst: dst}, nil
}

// Write adds p to the current batch. If the batch is full it is
// flushed to the underlying writer automatically.
func (w *Writer) Write(p []byte) (int, error) {
	if err := w.agg.Add(p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close flushes any remaining buffered chunks to the underlying writer.
func (w *Writer) Close() error {
	return w.agg.Flush()
}
