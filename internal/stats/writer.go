package stats

import "io"

// Writer wraps an io.Writer and records statistics for each Write call
// using the provided Collector.
type Writer struct {
	dst       io.Writer
	collector *Collector
}

// NewWriter creates a Writer that forwards writes to dst and records
// stats in the provided Collector.
func NewWriter(dst io.Writer, c *Collector) *Writer {
	return &Writer{dst: dst, collector: c}
}

// Write writes p to the underlying writer and records the chunk.
// If the underlying write fails, the chunk is recorded as dropped.
func (w *Writer) Write(p []byte) (int, error) {
	n, err := w.dst.Write(p)
	if err != nil {
		w.collector.RecordDrop(len(p))
		return n, err
	}
	w.collector.RecordChunk(n)
	return n, nil
}
