package transform_test

import (
	"bytes"
	"testing"

	"github.com/user/pipesnap/internal/transform"
)

func TestWriterPassthrough(t *testing.T) {
	var buf bytes.Buffer
	w := transform.NewWriter(&buf, transform.UpperCase)

	n, err := w.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Fatalf("expected n=5, got %d", n)
	}
	if buf.String() != "HELLO" {
		t.Fatalf("expected HELLO, got %q", buf.String())
	}
}

func TestWriterDropsChunk(t *testing.T) {
	var buf bytes.Buffer
	w := transform.NewWriter(&buf, transform.StripControl)

	// All control chars — should be dropped.
	n, err := w.Write([]byte("\x01\x02\x03"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Fatalf("expected n=3 (original length), got %d", n)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected empty buffer, got %q", buf.String())
	}
}

func TestWriterChain(t *testing.T) {
	var buf bytes.Buffer
	fn := transform.Chain(
		transform.LowerCase,
		transform.ReplaceAll("world", "pipesnap"),
	)
	w := transform.NewWriter(&buf, fn)

	_, err := w.Write([]byte("HELLO WORLD"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "hello pipesnap" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestWriterMultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	w := transform.NewWriter(&buf, transform.Truncate(3))

	for _, chunk := range []string{"abcdef", "xyz123", "hi"} {
		if _, err := w.Write([]byte(chunk)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if buf.String() != "abcxyzhi" {
		t.Fatalf("got %q", buf.String())
	}
}
