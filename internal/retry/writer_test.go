package retry

import (
	"errors"
	"io"
	"testing"
)

type closeTracker struct {
	*failWriter
	closed bool
}

func (c *closeTracker) Close() error {
	c.closed = true
	return nil
}

func TestNewWriterRejectsNilDst(t *testing.T) {
	_, err := NewWriter(nil, Config{MaxAttempts: 1})
	if err == nil {
		t.Fatal("expected error for nil dst")
	}
}

func TestWriterPassesDataThrough(t *testing.T) {
	fw := &failWriter{failFor: 0}
	w, err := NewWriter(fw, Config{MaxAttempts: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n, err := w.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	if n != 5 {
		t.Fatalf("expected n=5, got %d", n)
	}
	if string(fw.buf) != "hello" {
		t.Fatalf("unexpected buf: %q", fw.buf)
	}
}

func TestWriterReturnsErrAfterExhaustion(t *testing.T) {
	fw := &failWriter{failFor: 99}
	w, _ := NewWriter(fw, Config{MaxAttempts: 2})
	_, err := w.Write([]byte("x"))
	if !errors.Is(err, ErrMaxAttemptsExceeded) {
		t.Fatalf("expected ErrMaxAttemptsExceeded, got %v", err)
	}
	if w.Drops() != 1 {
		t.Fatalf("expected 1 drop, got %d", w.Drops())
	}
}

func TestWriterCloseCallsUnderlying(t *testing.T) {
	ct := &closeTracker{failWriter: &failWriter{}}
	w, _ := NewWriter(ct, Config{MaxAttempts: 1})
	if err := w.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if !ct.closed {
		t.Fatal("expected underlying closer to be called")
	}
}

func TestWriterCloseNoOpForNonCloser(t *testing.T) {
	w, _ := NewWriter(io.Discard, Config{MaxAttempts: 1})
	if err := w.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriterAttemptsTracked(t *testing.T) {
	fw := &failWriter{failFor: 1}
	w, _ := NewWriter(fw, Config{MaxAttempts: 3})
	w.Write([]byte("data")) //nolint:errcheck
	if w.Attempts() != 2 {
		t.Fatalf("expected 2 attempts, got %d", w.Attempts())
	}
}
