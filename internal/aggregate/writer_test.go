package aggregate

import (
	"bytes"
	"testing"
)

func TestNewWriterRejectsNilDst(t *testing.T) {
	_, err := NewWriter(nil, 2)
	if err == nil {
		t.Fatal("expected error for nil dst")
	}
}

func TestNewWriterRejectsZeroSize(t *testing.T) {
	_, err := NewWriter(&bytes.Buffer{}, 0)
	if err == nil {
		t.Fatal("expected error for size=0")
	}
}

func TestWriterAggregatesWrites(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, 3)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	for _, chunk := range []string{"foo", "bar", "baz"} {
		n, werr := w.Write([]byte(chunk))
		if werr != nil {
			t.Fatalf("Write: %v", werr)
		}
		if n != len(chunk) {
			t.Fatalf("expected n=%d, got %d", len(chunk), n)
		}
	}

	want := "foobarbaz"
	if buf.String() != want {
		t.Errorf("expected %q, got %q", want, buf.String())
	}
}

func TestWriterCloseFlushesRemainder(t *testing.T) {
	var buf bytes.Buffer
	w, _ := NewWriter(&buf, 4)

	_ , _ = w.Write([]byte("one"))
	_, _ = w.Write([]byte("two"))

	if buf.Len() != 0 {
		t.Fatal("expected no output before Close")
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if buf.String() != "onetwo" {
		t.Errorf("unexpected output after Close: %q", buf.String())
	}
}

func TestWriterReturnOriginalLen(t *testing.T) {
	var buf bytes.Buffer
	w, _ := NewWriter(&buf, 2)
	p := []byte("hello world")
	n, err := w.Write(p)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(p) {
		t.Errorf("expected n=%d, got %d", len(p), n)
	}
}
