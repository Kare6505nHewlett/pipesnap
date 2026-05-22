package burst

import (
	"bytes"
	"testing"
	"time"
)

func TestNewRejectsZeroCap(t *testing.T) {
	_, err := New(0, 1, time.Millisecond)
	if err == nil {
		t.Fatal("expected error for cap=0")
	}
}

func TestNewRejectsZeroRefill(t *testing.T) {
	_, err := New(1, 0, time.Millisecond)
	if err == nil {
		t.Fatal("expected error for refill=0")
	}
}

func TestNewRejectsZeroInterval(t *testing.T) {
	_, err := New(1, 1, 0)
	if err == nil {
		t.Fatal("expected error for interval=0")
	}
}

func TestAllowConsumesTokens(t *testing.T) {
	l, err := New(3, 1, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be able to call Allow 3 times immediately (full bucket).
	done := make(chan struct{})
	go func() {
		l.Allow()
		l.Allow()
		l.Allow()
		close(done)
	}()
	select {
	case <-done:
		// ok
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Allow blocked unexpectedly on burst capacity")
	}
}

func TestAllowBlocksWhenExhausted(t *testing.T) {
	// cap=1, slow refill — second Allow should block.
	l, err := New(1, 1, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	l.Allow() // consume the single token

	done := make(chan struct{})
	go func() {
		l.Allow()
		close(done)
	}()
	select {
	case <-done:
		t.Fatal("Allow should have blocked on empty bucket")
	case <-time.After(100 * time.Millisecond):
		// expected — still waiting for refill
	}
}

func TestRefillRestoresTokens(t *testing.T) {
	l, err := New(2, 2, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	l.Allow()
	l.Allow() // drain bucket

	time.Sleep(80 * time.Millisecond) // wait for refill

	done := make(chan struct{})
	go func() {
		l.Allow()
		close(done)
	}()
	select {
	case <-done:
		// ok — tokens were refilled
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Allow blocked after expected refill")
	}
}

func TestNewWriterPassesData(t *testing.T) {
	l, _ := New(10, 10, time.Millisecond)
	var buf bytes.Buffer
	w := NewWriter(&buf, l)

	payload := []byte("hello burst")
	n, err := w.Write(payload)
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if n != len(payload) {
		t.Fatalf("wrote %d bytes, want %d", n, len(payload))
	}
	if !bytes.Equal(buf.Bytes(), payload) {
		t.Fatalf("buffer mismatch: got %q, want %q", buf.Bytes(), payload)
	}
}
