package schema

import (
	"fmt"
	"io"
)

// Writer wraps an io.Writer and validates each Write call against a Schema.
// Chunks that fail validation are dropped; the original byte count is still
// returned so callers behave as if the write succeeded (matching the
// convention used by filter.Writer and transform.Writer).
type Writer struct {
	dst    io.Writer
	s      *Schema
	drops  int
	errors []error
}

// NewWriter returns a Writer that validates every chunk with s before
// forwarding it to dst.
func NewWriter(dst io.Writer, s *Schema) *Writer {
	return &Writer{dst: dst, s: s}
}

// Write validates p against the schema. If validation passes the chunk is
// forwarded to the underlying writer. If validation fails the chunk is
// silently dropped and the drop counter is incremented.
func (w *Writer) Write(p []byte) (int, error) {
	if err := w.s.Validate(p); err != nil {
		w.drops++
		w.errors = append(w.errors, fmt.Errorf("drop chunk: %w", err))
		return len(p), nil
	}
	n, err := w.dst.Write(p)
	if err != nil {
		return n, err
	}
	return len(p), nil
}

// Drops returns the number of chunks that failed schema validation.
func (w *Writer) Drops() int { return w.drops }

// Errors returns all validation errors encountered so far.
func (w *Writer) Errors() []error { return w.errors }
