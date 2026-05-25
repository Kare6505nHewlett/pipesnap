package skip_test

import (
	"bytes"
	"testing"

	"github.com/yourorg/pipesnap/internal/skip"
)

func TestNewRejectsNilDst(t *testing.T) {
	_, err := skip.New(nil, 1)
	if err == nil {
		t.Fatal("expected error for nil dst")
	}
}

func TestNewRejectsZeroN(t *testing.T) {
	_, err := skip.New(&bytes.Buffer{}, 0)
	if err == nil {
		t.Fatal("expected error for n=0")
	}
}

func TestNewRejectsNegativeN(t *testing.T) {
	_, err := skip.New(&bytes.Buffer{}, -3)
	if err == nil {
		t.Fatal("expected error for negative n")
	}
}

func TestSkipFirstChunk(t *testing.T) {
	var buf bytes.Buffer
	s, _ := skip.New(&buf, 1)

	s.Write([]byte("drop me"))
	s.Write([]byte("keep me"))

	if buf.String() != "keep me" {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestSkipMultipleChunks(t *testing.T) {
	var buf bytes.Buffer
	s, _ := skip.New(&buf, 3)

	chunks := []string{"a", "b", "c", "d", "e"}
	for _, c := range chunks {
		s.Write([]byte(c))
	}

	if buf.String() != "de" {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestCounters(t *testing.T) {
	var buf bytes.Buffer
	s, _ := skip.New(&buf, 2)

	for i := 0; i < 5; i++ {
		s.Write([]byte("x"))
	}

	if s.Seen() != 5 {
		t.Errorf("Seen: want 5, got %d", s.Seen())
	}
	if s.Dropped() != 2 {
		t.Errorf("Dropped: want 2, got %d", s.Dropped())
	}
	if s.Passed() != 3 {
		t.Errorf("Passed: want 3, got %d", s.Passed())
	}
}

func TestSkipMoreThanAvailable(t *testing.T) {
	var buf bytes.Buffer
	s, _ := skip.New(&buf, 10)

	for i := 0; i < 3; i++ {
		s.Write([]byte("x"))
	}

	if buf.Len() != 0 {
		t.Errorf("expected empty buffer, got %q", buf.String())
	}
	if s.Dropped() != 3 {
		t.Errorf("Dropped: want 3, got %d", s.Dropped())
	}
}

func TestNewWriterPanicsOnBadArgs(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	skip.NewWriter(nil, 1)
}

func TestWriteReturnsOriginalLen(t *testing.T) {
	var buf bytes.Buffer
	s, _ := skip.New(&buf, 2)

	p := []byte("hello")
	n, err := s.Write(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(p) {
		t.Errorf("Write returned %d, want %d", n, len(p))
	}
}
