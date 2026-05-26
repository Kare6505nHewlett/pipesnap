package cap_test

import (
	"bytes"
	"testing"

	"github.com/yourorg/pipesnap/internal/cap"
)

func TestNewRejectsNegativeMaxChunks(t *testing.T) {
	_, err := cap.New(-1, 0)
	if err == nil {
		t.Fatal("expected error for negative maxChunks")
	}
}

func TestNewRejectsNegativeMaxBytes(t *testing.T) {
	_, err := cap.New(0, -1)
	if err == nil {
		t.Fatal("expected error for negative maxBytes")
	}
}

func TestNewRejectsBothZero(t *testing.T) {
	_, err := cap.New(0, 0)
	if err == nil {
		t.Fatal("expected error when both limits are zero")
	}
}

func TestAllowByChunks(t *testing.T) {
	l, err := cap.New(3, 0)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if !l.Allow(10) {
			t.Fatalf("chunk %d should be allowed", i)
		}
	}
	if l.Allow(10) {
		t.Fatal("chunk 4 should be rejected")
	}
	if l.Chunks() != 3 {
		t.Fatalf("expected 3 chunks, got %d", l.Chunks())
	}
}

func TestAllowByBytes(t *testing.T) {
	l, err := cap.New(0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if !l.Allow(15) {
		t.Fatal("first chunk should be allowed")
	}
	// 15+10 = 25 > 20 → reject
	if l.Allow(10) {
		t.Fatal("second chunk should be rejected")
	}
	if l.Bytes() != 15 {
		t.Fatalf("expected 15 bytes, got %d", l.Bytes())
	}
}

func TestExhausted(t *testing.T) {
	l, _ := cap.New(1, 0)
	if l.Exhausted() {
		t.Fatal("should not be exhausted before any write")
	}
	l.Allow(5)
	if !l.Exhausted() {
		t.Fatal("should be exhausted after reaching chunk cap")
	}
}

func TestWriterForwardsUntilCap(t *testing.T) {
	var buf bytes.Buffer
	w, err := cap.NewWriter(&buf, 2, 0)
	if err != nil {
		t.Fatal(err)
	}

	w.Write([]byte("hello"))
	w.Write([]byte("world"))
	w.Write([]byte("dropped")) // should be silently dropped

	if buf.String() != "helloworld" {
		t.Fatalf("unexpected buffer content: %q", buf.String())
	}
	if w.Limiter().Chunks() != 2 {
		t.Fatalf("expected 2 chunks forwarded, got %d", w.Limiter().Chunks())
	}
}

func TestWriterReturnOriginalLen(t *testing.T) {
	var buf bytes.Buffer
	w, _ := cap.NewWriter(&buf, 1, 0)
	w.Write([]byte("first")) // allowed
	n, err := w.Write([]byte("second")) // dropped
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len("second") {
		t.Fatalf("expected %d, got %d", len("second"), n)
	}
}
