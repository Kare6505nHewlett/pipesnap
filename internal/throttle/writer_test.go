package throttle

import (
	"bytes"
	"io"
	"testing"
	"time"
)

func TestWriterPassesAllWhenBudgetSufficient(t *testing.T) {
	th, _ := New(10, time.Hour)
	var buf bytes.Buffer
	w := NewWriter(&buf, th)

	payloads := [][]byte{[]byte("hello"), []byte(" "), []byte("world")}
	for _, p := range payloads {
		if _, err := w.Write(p); err != nil {
			t.Fatal(err)
		}
	}
	if got := buf.String(); got != "hello world" {
		t.Fatalf("want %q, got %q", "hello world", got)
	}
}

func TestWriterDropsAfterBudgetExhausted(t *testing.T) {
	th, _ := New(2, time.Hour)
	var buf bytes.Buffer
	w := NewWriter(&buf, th)

	for i := 0; i < 5; i++ {
		n, err := w.Write([]byte("x"))
		if err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Fatalf("want n=1, got %d", n)
		}
	}
	// Only 2 chunks should have reached the underlying writer.
	if got := buf.Len(); got != 2 {
		t.Fatalf("want 2 bytes written, got %d", got)
	}
}

func TestWriterReturnOriginalLen(t *testing.T) {
	th, _ := New(1, time.Hour)
	th.Allow() // exhaust budget

	w := NewWriter(io.Discard, th)
	data := []byte("dropped chunk")
	n, err := w.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Fatalf("want len=%d, got %d", len(data), n)
	}
}

func TestWriterResumesAfterWindowReset(t *testing.T) {
	th, _ := New(1, 20*time.Millisecond)
	var buf bytes.Buffer
	w := NewWriter(&buf, th)

	w.Write([]byte("a")) // consumes budget
	w.Write([]byte("b")) // dropped

	time.Sleep(30 * time.Millisecond)
	w.Write([]byte("c")) // new window, should pass

	if got := buf.String(); got != "ac" {
		t.Fatalf("want %q, got %q", "ac", got)
	}
}
