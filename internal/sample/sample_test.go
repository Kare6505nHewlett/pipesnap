package sample

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewRejectsZero(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for n=0")
	}
}

func TestNewRejectsNegative(t *testing.T) {
	_, err := New(-3)
	if err == nil {
		t.Fatal("expected error for n=-3")
	}
}

func TestSampleEveryOne(t *testing.T) {
	s, _ := New(1)
	for i := 0; i < 5; i++ {
		if !s.Sample() {
			t.Fatalf("n=1: chunk %d should always pass", i)
		}
	}
	if s.Dropped() != 0 {
		t.Fatalf("expected 0 dropped, got %d", s.Dropped())
	}
}

func TestSampleEveryThird(t *testing.T) {
	s, _ := New(3)
	results := make([]bool, 9)
	for i := range results {
		results[i] = s.Sample()
	}
	// chunks 3, 6, 9 (1-indexed) should pass
	expected := []bool{false, false, true, false, false, true, false, false, true}
	for i, want := range expected {
		if results[i] != want {
			t.Errorf("chunk %d: got %v want %v", i+1, results[i], want)
		}
	}
	if s.Dropped() != 6 {
		t.Fatalf("expected 6 dropped, got %d", s.Dropped())
	}
	if s.Seen() != 9 {
		t.Fatalf("expected 9 seen, got %d", s.Seen())
	}
}

func TestNewWriterInvalidN(t *testing.T) {
	_, err := NewWriter(&bytes.Buffer{}, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewWriterPassesEverySecond(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	chunks := []string{"alpha", "beta", "gamma", "delta"}
	for _, c := range chunks {
		n, werr := w.Write([]byte(c))
		if werr != nil {
			t.Fatalf("write error: %v", werr)
		}
		if n != len(c) {
			t.Fatalf("short write reported")
		}
	}
	// chunks 2 and 4 (beta, delta) should be written
	got := buf.String()
	if !strings.Contains(got, "beta") || !strings.Contains(got, "delta") {
		t.Errorf("expected beta and delta in output, got %q", got)
	}
	if strings.Contains(got, "alpha") || strings.Contains(got, "gamma") {
		t.Errorf("expected alpha and gamma to be dropped, got %q", got)
	}
}
