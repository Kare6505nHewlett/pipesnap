package cap

import "io"

// Writer forwards chunks to an underlying io.Writer until the cap is reached.
// After the cap is reached, Write returns the original length without error so
// that callers that ignore drop semantics continue to function correctly.
type Writer struct {
	dst io.Writer
	lim *Limiter
}

// Write forwards p to the underlying writer if the limiter allows it.
// Once the cap is reached the data is silently dropped but the full length
// is still returned so pipeline callers do not see spurious errors.
func (w *Writer) Write(p []byte) (int, error) {
	if !w.lim.Allow(len(p)) {
		return len(p), nil
	}
	return w.dst.Write(p)
}

// Close closes the underlying writer if it implements io.Closer.
func (w *Writer) Close() error {
	if c, ok := w.dst.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Limiter exposes the underlying Limiter for inspection.
func (w *Writer) Limiter() *Limiter { return w.lim }
