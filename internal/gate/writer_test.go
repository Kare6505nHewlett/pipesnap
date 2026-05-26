package gate

import (
	"bytes"
	"io"
	"testing"
)

// nopCloser wraps a *bytes.Buffer to satisfy io.WriteCloser.
type nopCloser struct{ *bytes.Buffer }

func (n *nopCloser) Close() error { return nil }

type trackCloser struct {
	*bytes.Buffer
	closed bool
}

func (t *trackCloser) Close() error { t.closed = true; return nil }

func TestNewWriterRejectsNilGate(t *testing.T) {
	_, err := NewWriter(nil, &nopCloser{bytes.NewBuffer(nil)})
	if err == nil {
		t.Fatal("expected error for nil gate")
	}
}

func TestNewWriterRejectsNilDst(t *testing.T) {
	buf := &bytes.Buffer{}
	g, _ := New(buf, func([]byte) bool { return true })
	_, err := NewWriter(g, nil)
	if err == nil {
		t.Fatal("expected error for nil dst")
	}
}

func TestWriterPassesData(t *testing.T) {
	buf := &bytes.Buffer{}
	g, _ := New(buf, func([]byte) bool { return true })
	g.Open()

	w, err := NewWriter(g, &nopCloser{bytes.NewBuffer(nil)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w.Write([]byte("hello"))
	if buf.String() != "hello" {
		t.Fatalf("expected 'hello', got %q", buf.String())
	}
}

func TestWriterCloseCallsUnderlying(t *testing.T) {
	buf := &bytes.Buffer{}
	g, _ := New(buf, func([]byte) bool { return false })
	track := &trackCloser{Buffer: bytes.NewBuffer(nil)}

	w, _ := NewWriter(g, track)
	w.Close()

	if !track.closed {
		t.Fatal("expected underlying closer to be called")
	}
}

func TestWriterGateAccessor(t *testing.T) {
	buf := &bytes.Buffer{}
	g, _ := New(buf, func([]byte) bool { return false })
	w, _ := NewWriter(g, &nopCloser{bytes.NewBuffer(nil)})

	if w.Gate() != g {
		t.Fatal("Gate() should return the same gate passed to NewWriter")
	}
}

func TestWriterImplementsWriteCloser(t *testing.T) {
	buf := &bytes.Buffer{}
	g, _ := New(buf, func([]byte) bool { return true })
	w, _ := NewWriter(g, &nopCloser{bytes.NewBuffer(nil)})

	var _ io.WriteCloser = w
}
