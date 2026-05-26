package gate

import (
	"bytes"
	"errors"
	"testing"
)

func TestNewRejectsNilDst(t *testing.T) {
	_, err := New(nil, func([]byte) bool { return true })
	if err == nil {
		t.Fatal("expected error for nil dst")
	}
}

func TestNewRejectsNilPred(t *testing.T) {
	_, err := New(&bytes.Buffer{}, nil)
	if err == nil {
		t.Fatal("expected error for nil pred")
	}
}

func TestGateStartsClosed(t *testing.T) {
	buf := &bytes.Buffer{}
	g, _ := New(buf, func([]byte) bool { return false })

	if g.IsOpen() {
		t.Fatal("gate should start closed")
	}

	g.Write([]byte("hello"))
	if buf.Len() != 0 {
		t.Fatalf("expected no bytes forwarded, got %d", buf.Len())
	}
}

func TestGateOpensOnMatchingChunk(t *testing.T) {
	buf := &bytes.Buffer{}
	g, _ := New(buf, func(p []byte) bool {
		return bytes.Contains(p, []byte("START"))
	})

	g.Write([]byte("noise"))
	g.Write([]byte("START"))
	g.Write([]byte("after"))

	if !g.IsOpen() {
		t.Fatal("gate should be open after trigger chunk")
	}
	if got := buf.String(); got != "STARTafter" {
		t.Fatalf("unexpected forwarded data: %q", got)
	}
}

func TestGateForceOpen(t *testing.T) {
	buf := &bytes.Buffer{}
	g, _ := New(buf, func([]byte) bool { return false })

	g.Open()
	g.Write([]byte("data"))

	if buf.String() != "data" {
		t.Fatalf("expected data forwarded after force-open")
	}
}

func TestGateForceClose(t *testing.T) {
	buf := &bytes.Buffer{}
	g, _ := New(buf, func([]byte) bool { return true })

	g.Open()
	g.Close()
	g.Write([]byte("blocked"))

	if buf.Len() != 0 {
		t.Fatalf("expected no bytes after force-close")
	}
}

func TestSeenAndPassedCounters(t *testing.T) {
	buf := &bytes.Buffer{}
	calls := 0
	g, _ := New(buf, func(p []byte) bool {
		calls++
		return calls >= 3
	})

	for i := 0; i < 5; i++ {
		g.Write([]byte("x"))
	}

	if g.Seen() != 5 {
		t.Fatalf("expected 5 seen, got %d", g.Seen())
	}
	// gate opens on 3rd write and stays open → 3 passed
	if g.Passed() != 3 {
		t.Fatalf("expected 3 passed, got %d", g.Passed())
	}
}

func TestWriteReturnsErrorFromDst(t *testing.T) {
	w := &errWriter{}
	g, _ := New(w, func([]byte) bool { return true })
	g.Open()

	_, err := g.Write([]byte("boom"))
	if err == nil {
		t.Fatal("expected error from dst")
	}
}

type errWriter struct{}

func (e *errWriter) Write([]byte) (int, error) {
	return 0, errors.New("write error")
}
