package audit

import (
	"fmt"
	"io"
)

// Writer wraps a destination io.Writer and audits every chunk written to it.
type Writer struct {
	dst    io.Writer
	logger *Logger
}

// NewWriter returns a Writer that forwards each Write to dst and records an
// audit entry via logger.
func NewWriter(dst io.Writer, logger *Logger) (*Writer, error) {
	if dst == nil {
		return nil, fmt.Errorf("audit: writer dst must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("audit: writer logger must not be nil")
	}
	return &Writer{dst: dst, logger: logger}, nil
}

// Write audits data and then forwards it to the underlying writer.
// The original length is always returned so upstream writers are not confused
// by audit-only errors.
func (w *Writer) Write(data []byte) (int, error) {
	// Best-effort audit — do not block the pipeline on log failures.
	_ = w.logger.Record(data)
	return w.dst.Write(data)
}

// Close closes the underlying writer if it implements io.Closer.
func (w *Writer) Close() error {
	if c, ok := w.dst.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
