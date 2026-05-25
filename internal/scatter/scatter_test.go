package scatter

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestNewRejectsNoDests(t *testing.T) {
	_, err := New(RoundRobin)
	if err == nil {
		t.Fatal("expected error for zero destinations")
	}
}

func TestNewRejectsNilDest(t *testing.T) {
	_, err := New(RoundRobin, &bytes.Buffer{}, nil)
	if err == nil {
		t.Fatal("expected error for nil destination")
	}
}

func TestRoundRobinDistributes(t *testing.T) {
	var a, b, c bytes.Buffer
	s, err := New(RoundRobin, &a, &b, &c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	chunks := [][]byte{[]byte("one"), []byte("two"), []byte("three"), []byte("four")}
	for _, ch := range chunks {
		if _, err := s.Write(ch); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	if a.String() != "onefour" {
		t.Errorf("dest a: got %q, want %q", a.String(), "onefour")
	}
	if b.String() != "two" {
		t.Errorf("dest b: got %q, want %q", b.String(), "two")
	}
	if c.String() != "three" {
		t.Errorf("dest c: got %q, want %q", c.String(), "three")
	}
}

func TestHashBasedIsDeterministic(t *testing.T) {
	var a, b bytes.Buffer
	s1, _ := New(HashBased, &a, &b)
	var c, d bytes.Buffer
	s2, _ := New(HashBased, &c, &d)

	chunk := []byte("hello")
	s1.Write(chunk)
	s2.Write(chunk)

	if a.String() != c.String() || b.String() != d.String() {
		t.Error("hash-based scatter is not deterministic for same input")
	}
}

func TestHashBasedSpreads(t *testing.T) {
	var a, b bytes.Buffer
	s, _ := New(HashBased, &a, &b)
	chunks := [][]byte{
		[]byte("alpha"), []byte("beta"), []byte("gamma"), []byte("delta"),
		[]byte("epsilon"), []byte("zeta"), []byte("eta"), []byte("theta"),
	}
	for _, ch := range chunks {
		s.Write(ch)
	}
	if a.Len() == 0 || b.Len() == 0 {
		t.Error("expected hash-based scatter to write to both destinations")
	}
}

func TestWriteReturnsErrorOnDestFailure(t *testing.T) {
	w := &errWriter{err: errors.New("boom")}
	s, _ := New(RoundRobin, w)
	_, err := s.Write([]byte("data"))
	if err == nil {
		t.Fatal("expected error from failing destination")
	}
}

func TestLen(t *testing.T) {
	s, _ := New(RoundRobin, &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	if s.Len() != 3 {
		t.Errorf("Len: got %d, want 3", s.Len())
	}
}

type errWriter struct{ err error }

func (e *errWriter) Write(_ []byte) (int, error) { return 0, e.err }

var _ io.Writer = (*errWriter)(nil)
