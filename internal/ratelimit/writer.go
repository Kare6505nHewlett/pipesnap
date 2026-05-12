package ratelimit

import "io"

// Writer wraps an io.Writer and applies rate limiting before each Write call.
type Writer struct {
	w       io.Writer
	limiter *Limiter
}

// NewWriter returns a Writer that rate-limits writes to w using l.
func NewWriter(w io.Writer, l *Limiter) *Writer {
	return &Writer{w: w, limiter: l}
}

// Write waits according to the limiter policy and then writes p to the
// underlying writer. It returns the number of bytes written and any error.
func (rw *Writer) Write(p []byte) (int, error) {
	rw.limiter.Wait(len(p))
	return rw.w.Write(p)
}
