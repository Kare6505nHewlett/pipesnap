package schema

import (
	"bytes"
	"testing"
)

func makeSchema(t *testing.T) *Schema {
	t.Helper()
	s, err := New([]Field{
		{Name: "msg", Type: TypeString, Required: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestWriterPassthrough(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, makeSchema(t))
	chunk := []byte(`{"msg":"hello"}`)
	n, err := w.Write(chunk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(chunk) {
		t.Fatalf("expected n=%d, got %d", len(chunk), n)
	}
	if buf.String() != string(chunk) {
		t.Fatalf("unexpected buf: %q", buf.String())
	}
	if w.Drops() != 0 {
		t.Fatalf("expected 0 drops, got %d", w.Drops())
	}
}

func TestWriterDropsInvalidChunk(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, makeSchema(t))
	chunk := []byte(`{"msg":42}`) // msg should be string
	n, err := w.Write(chunk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(chunk) {
		t.Fatalf("expected n=%d, got %d", len(chunk), n)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected empty buf, got %q", buf.String())
	}
	if w.Drops() != 1 {
		t.Fatalf("expected 1 drop, got %d", w.Drops())
	}
	if len(w.Errors()) != 1 {
		t.Fatalf("expected 1 error, got %d", len(w.Errors()))
	}
}

func TestWriterMultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, makeSchema(t))
	chunks := [][]byte{
		[]byte(`{"msg":"a"}`),
		[]byte(`{"msg":99}`),
		[]byte(`{"msg":"c"}`),
	}
	for _, c := range chunks {
		if _, err := w.Write(c); err != nil {
			t.Fatal(err)
		}
	}
	if w.Drops() != 1 {
		t.Fatalf("expected 1 drop, got %d", w.Drops())
	}
}
