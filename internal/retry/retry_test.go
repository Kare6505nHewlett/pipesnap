package retry

import (
	"errors"
	"io"
	"testing"
)

// failWriter fails the first n writes then succeeds.
type failWriter struct {
	failFor int
	calls   int
	buf     []byte
}

func (f *failWriter) Write(p []byte) (int, error) {
	f.calls++
	if f.calls <= f.failFor {
		return 0, errors.New("transient error")
	}
	f.buf = append(f.buf, p...)
	return len(p), nil
}

func TestNewRejectsNilDst(t *testing.T) {
	_, err := New(nil, Config{MaxAttempts: 3})
	if err == nil {
		t.Fatal("expected error for nil dst")
	}
}

func TestNewRejectsZeroAttempts(t *testing.T) {
	_, err := New(io.Discard, Config{MaxAttempts: 0})
	if err == nil {
		t.Fatal("expected error for MaxAttempts=0")
	}
}

func TestSuccessOnFirstAttempt(t *testing.T) {
	fw := &failWriter{failFor: 0}
	r, _ := New(fw, Config{MaxAttempts: 3})
	_, err := r.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Attempts() != 1 {
		t.Fatalf("expected 1 attempt, got %d", r.Attempts())
	}
	if r.Drops() != 0 {
		t.Fatalf("expected 0 drops, got %d", r.Drops())
	}
}

func TestSuccessAfterRetry(t *testing.T) {
	fw := &failWriter{failFor: 2}
	r, _ := New(fw, Config{MaxAttempts: 5})
	_, err := r.Write([]byte("data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Attempts() != 3 {
		t.Fatalf("expected 3 attempts, got %d", r.Attempts())
	}
	if string(fw.buf) != "data" {
		t.Fatalf("unexpected buf: %q", fw.buf)
	}
}

func TestExhaustsAllAttempts(t *testing.T) {
	fw := &failWriter{failFor: 10}
	r, _ := New(fw, Config{MaxAttempts: 3})
	_, err := r.Write([]byte("x"))
	if !errors.Is(err, ErrMaxAttemptsExceeded) {
		t.Fatalf("expected ErrMaxAttemptsExceeded, got %v", err)
	}
	if r.Drops() != 1 {
		t.Fatalf("expected 1 drop, got %d", r.Drops())
	}
	if r.Attempts() != 3 {
		t.Fatalf("expected 3 attempts, got %d", r.Attempts())
	}
}

func TestDropsAccumulate(t *testing.T) {
	fw := &failWriter{failFor: 100}
	r, _ := New(fw, Config{MaxAttempts: 1})
	for i := 0; i < 4; i++ {
		r.Write([]byte("chunk")) //nolint:errcheck
	}
	if r.Drops() != 4 {
		t.Fatalf("expected 4 drops, got %d", r.Drops())
	}
}
