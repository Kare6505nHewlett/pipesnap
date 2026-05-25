package join

import "io"

// Writer is an io.WriteCloser that feeds each Write payload into a Joiner and
// forwards joined chunks to the underlying destination.
type Writer struct {
	j *Joiner
}

// NewWriter wraps the provided Joiner in a Writer so it can be used anywhere
// an io.WriteCloser is accepted.
func NewWriter(j *Joiner) *Writer {
	if j == nil {
		panic("join: NewWriter called with nil Joiner")
	}
	return &Writer{j: j}
}

// Write adds p to the joiner's buffer. The length of p is always returned so
// that callers that wrap this writer see no short-write errors even when the
// joiner is still accumulating.
func (w *Writer) Write(p []byte) (int, error) {
	if err := w.j.Add(p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close flushes any remaining buffered chunks and returns any error from the
// underlying emit function.
func (w *Writer) Close() error {
	return w.j.Flush()
}

// Len returns the number of chunks currently buffered by the underlying joiner.
func (w *Writer) Len() int { return w.j.Len() }

// Ensure Writer satisfies io.WriteCloser at compile time.
var _ io.WriteCloser = (*Writer)(nil)
