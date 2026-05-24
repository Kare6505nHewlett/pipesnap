package expand

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewRejectsEmptyDelimiter(t *testing.T) {
	_, err := New([]byte{})
	if err == nil {
		t.Fatal("expected error for empty delimiter")
	}
}

func TestNewAcceptsValidDelimiter(t *testing.T) {
	_, err := New([]byte("\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplySingleSegment(t *testing.T) {
	e, _ := New([]byte("\n"))
	var buf bytes.Buffer
	n, err := e.Apply(&buf, []byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Fatalf("expected n=5, got %d", n)
	}
	if buf.String() != "hello" {
		t.Fatalf("expected 'hello', got %q", buf.String())
	}
}

func TestApplyMultipleSegments(t *testing.T) {
	e, _ := New([]byte("|"))
	var calls []string
	w := &callWriter{fn: func(p []byte) { calls = append(calls, string(p)) }}
	e.Apply(w, []byte("a|b|c"))
	if len(calls) != 3 {
		t.Fatalf("expected 3 writes, got %d", len(calls))
	}
	if calls[0] != "a" || calls[1] != "b" || calls[2] != "c" {
		t.Fatalf("unexpected segments: %v", calls)
	}
}

func TestApplyDropsEmptySegments(t *testing.T) {
	e, _ := New([]byte("\n"))
	var calls []string
	w := &callWriter{fn: func(p []byte) { calls = append(calls, string(p)) }}
	e.Apply(w, []byte("a\n\nb"))
	if len(calls) != 2 {
		t.Fatalf("expected 2 writes (empty segment dropped), got %d", len(calls))
	}
}

func TestApplyReturnsOriginalLen(t *testing.T) {
	e, _ := New([]byte(","))
	var buf bytes.Buffer
	input := []byte("x,y,z")
	n, _ := e.Apply(&buf, input)
	if n != len(input) {
		t.Fatalf("expected n=%d, got %d", len(input), n)
	}
}

func TestNewWriterRejectsEmptyDelimiter(t *testing.T) {
	_, err := NewWriter(&bytes.Buffer{}, []byte{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewWriterExpandsOnWrite(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, []byte("\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer w.Close()
	w.Write([]byte("line1\nline2\nline3"))
	got := buf.String()
	if !strings.Contains(got, "line1") || !strings.Contains(got, "line2") || !strings.Contains(got, "line3") {
		t.Fatalf("unexpected output: %q", got)
	}
}

// callWriter records each Write call via fn.
type callWriter struct{ fn func([]byte) }

func (c *callWriter) Write(p []byte) (int, error) {
	c.fn(p)
	return len(p), nil
}
