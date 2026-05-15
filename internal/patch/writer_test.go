package patch

import (
	"bytes"
	"testing"
)

func TestWriterPassthrough(t *testing.T) {
	p, _ := New([]Sub{{Find: []byte("x"), Replace: []byte("y")}})
	var buf bytes.Buffer
	w := NewWriter(&buf, p)
	n, err := w.Write([]byte("axb"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Fatalf("expected n=3, got %d", n)
	}
	if buf.String() != "ayb" {
		t.Fatalf("got %q, want %q", buf.String(), "ayb")
	}
}

func TestWriterMultipleWrites(t *testing.T) {
	p, _ := New([]Sub{{Find: []byte("go"), Replace: []byte("GO")}})
	var buf bytes.Buffer
	w := NewWriter(&buf, p)
	w.Write([]byte("go lang "))
	w.Write([]byte("is go"))
	want := "GO lang is GO"
	if buf.String() != want {
		t.Fatalf("got %q, want %q", buf.String(), want)
	}
}

func TestWriterReturnOriginalLen(t *testing.T) {
	// Even when the patched output is longer, n should equal len(input).
	p, _ := New([]Sub{{Find: []byte("a"), Replace: []byte("aaa")}})
	var buf bytes.Buffer
	w := NewWriter(&buf, p)
	n, err := w.Write([]byte("ab"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected n=2 (original len), got %d", n)
	}
	if buf.String() != "aaab" {
		t.Fatalf("got %q, want %q", buf.String(), "aaab")
	}
}
