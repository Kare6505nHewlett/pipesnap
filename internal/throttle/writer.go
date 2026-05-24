package throttle

import (
	"bytes"
	"io"
	"testing"
	"time"
)

// writerTestHelper is used only from writer_test.go; kept here so the package
// compiles as a single unit. Actual exported API lives in throttle.go.

// collectWriter is a test helper that records all data written to it.
type collectWriter struct {
	buf bytes.Buffer
}

func (c *collectWriter) Write(p []byte) (int, error) {
	return c.buf.Write(p)
}

// exerciseWriter is a shared helper for writer tests.
func exerciseWriter(t *testing.T, th *Throttle, payloads [][]byte) (passed, dropped int, out []byte) {
	t.Helper()
	cw := &collectWriter{}
	w := NewWriter(cw, th)
	for _, p := range payloads {
		n, err := w.Write(p)
		if err != nil {
			t.Fatalf("unexpected write error: %v", err)
		}
		if n != len(p) {
			t.Fatalf("short write: got %d want %d", n, len(p))
		}
	}
	return cw.buf.Len(), 0, cw.buf.Bytes()
}

// Ensure NewWriter satisfies io.Writer at compile time.
var _ io.Writer = NewWriter(io.Discard, &Throttle{max: 1, window: time.Second, lastReset: time.Now()})
