package dedupe

import (
	"bytes"
	"testing"
)

func TestNewFilterIsEmpty(t *testing.T) {
	f := New()
	if f.Len() != 0 {
		t.Fatalf("expected 0 seen, got %d", f.Len())
	}
}

func TestFirstChunkNotDuplicate(t *testing.T) {
	f := New()
	if f.IsDuplicate([]byte("hello")) {
		t.Fatal("first occurrence should not be a duplicate")
	}
}

func TestSecondChunkIsDuplicate(t *testing.T) {
	f := New()
	f.IsDuplicate([]byte("hello"))
	if !f.IsDuplicate([]byte("hello")) {
		t.Fatal("second occurrence should be a duplicate")
	}
}

func TestDistinctChunksNotDuplicate(t *testing.T) {
	f := New()
	f.IsDuplicate([]byte("hello"))
	if f.IsDuplicate([]byte("world")) {
		t.Fatal("distinct chunk should not be a duplicate")
	}
}

func TestLenTracksUnique(t *testing.T) {
	f := New()
	f.IsDuplicate([]byte("a"))
	f.IsDuplicate([]byte("b"))
	f.IsDuplicate([]byte("a")) // duplicate
	if f.Len() != 2 {
		t.Fatalf("expected 2 unique, got %d", f.Len())
	}
}

func TestResetClearsSeen(t *testing.T) {
	f := New()
	f.IsDuplicate([]byte("hello"))
	f.Reset()
	if f.Len() != 0 {
		t.Fatalf("expected 0 after reset, got %d", f.Len())
	}
	if f.IsDuplicate([]byte("hello")) {
		t.Fatal("after reset chunk should not be duplicate")
	}
}

func TestWriterPassesUnique(t *testing.T) {
	var buf bytes.Buffer
	f := New()
	w := NewWriter(&buf, f)

	w.Write([]byte("chunk1"))
	w.Write([]byte("chunk2"))

	if buf.String() != "chunk1chunk2" {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestWriterDropsDuplicates(t *testing.T) {
	var buf bytes.Buffer
	f := New()
	w := NewWriter(&buf, f)

	w.Write([]byte("hello"))
	w.Write([]byte("hello")) // should be dropped
	w.Write([]byte("world"))

	if buf.String() != "helloworld" {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestWriterReturnsFakeNOnDrop(t *testing.T) {
	var buf bytes.Buffer
	f := New()
	w := NewWriter(&buf, f)

	w.Write([]byte("dup"))
	n, err := w.Write([]byte("dup"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Fatalf("expected n=3 on drop, got %d", n)
	}
}
