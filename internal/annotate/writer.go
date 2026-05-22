package annotate

import (
	"fmt"
	"io"
)

// Writer wraps an io.Writer and annotates each written chunk before
// forwarding it downstream. Chunks that are not valid JSON objects are
// dropped and an error is returned to the caller.
type Writer struct {
	dst  io.Writer
	ann  *Annotator
	drops int
}

// NewWriter returns a Writer that annotates every chunk with ann before
// writing to dst.
func NewWriter(dst io.Writer, ann *Annotator) (*Writer, error) {
	if dst == nil {
		return nil, fmt.Errorf("annotate: destination writer must not be nil")
	}
	if ann == nil {
		return nil, fmt.Errorf("annotate: annotator must not be nil")
	}
	return &Writer{dst: dst, ann: ann}, nil
}

// Write annotates p and writes the result to the underlying writer.
// It returns the original length of p on success so callers that track
// bytes written see consistent counts.
func (w *Writer) Write(p []byte) (int, error) {
	annotated, err := w.ann.Apply(p)
	if err != nil {
		w.drops++
		return 0, err
	}
	if _, err := w.dst.Write(annotated); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Drops returns the number of chunks dropped due to annotation errors.
func (w *Writer) Drops() int { return w.drops }
