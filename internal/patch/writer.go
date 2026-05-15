package patch

import "io"

// Writer wraps an io.Writer and applies a Patcher to every chunk before
// forwarding it to the underlying writer.
type Writer struct {
	w io.Writer
	p *Patcher
}

// NewWriter returns a Writer that patches each Write call using p before
// passing the result to w.
func NewWriter(w io.Writer, p *Patcher) *Writer {
	return &Writer{w: w, p: p}
}

// Write applies the patcher to b and writes the transformed bytes to the
// underlying writer. The returned n reflects the length of the original
// input so callers satisfy the io.Writer contract.
func (pw *Writer) Write(b []byte) (int, error) {
	patched := pw.p.Apply(b)
	_, err := pw.w.Write(patched)
	return len(b), err
}
