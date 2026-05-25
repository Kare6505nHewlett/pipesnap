package join

import (
	"bytes"
	"errors"
	"testing"
)

func TestNewRejectsZeroMaxChunks(t *testing.T) {
	_, err := New([]byte(","), 0, func([]byte) error { return nil })
	if err == nil {
		t.Fatal("expected error for maxChunks=0")
	}
}

func TestNewRejectsNilOnEmit(t *testing.T) {
	_, err := New([]byte(","), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil onEmit")
	}
}

func TestFlushNoOpWhenEmpty(t *testing.T) {
	called := false
	j, _ := New([]byte(","), 4, func([]byte) error {
		called = true
		return nil
	})
	if err := j.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("onEmit should not be called for empty flush")
	}
}

func TestAddEmptyChunkIsIgnored(t *testing.T) {
	j, _ := New([]byte(","), 4, func([]byte) error { return nil })
	_ = j.Add([]byte{})
	if j.Len() != 0 {
		t.Fatalf("expected 0 buffered, got %d", j.Len())
	}
}

func TestFlushJoinsWithDelimiter(t *testing.T) {
	var got []byte
	j, _ := New([]byte("|"), 10, func(p []byte) error {
		got = append([]byte{}, p...)
		return nil
	})
	_ = j.Add([]byte("foo"))
	_ = j.Add([]byte("bar"))
	_ = j.Add([]byte("baz"))
	if err := j.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "foo|bar|baz"
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAutoFlushOnMaxChunks(t *testing.T) {
	var results []string
	j, _ := New([]byte("-"), 2, func(p []byte) error {
		results = append(results, string(p))
		return nil
	})
	_ = j.Add([]byte("a"))
	_ = j.Add([]byte("b")) // triggers auto-flush
	_ = j.Add([]byte("c"))
	_ = j.Flush()

	if len(results) != 2 {
		t.Fatalf("expected 2 emits, got %d", len(results))
	}
	if results[0] != "a-b" {
		t.Errorf("first emit: got %q, want %q", results[0], "a-b")
	}
	if results[1] != "c" {
		t.Errorf("second emit: got %q, want %q", results[1], "c")
	}
}

func TestWriterJoinsAndCloseFlushes(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, []byte(" "), 3)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	for _, s := range []string{"hello", "world"} {
		n, werr := w.Write([]byte(s))
		if werr != nil {
			t.Fatalf("Write: %v", werr)
		}
		if n != len(s) {
			t.Fatalf("short write: got %d, want %d", n, len(s))
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if got := buf.String(); got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestWriterPropagatesEmitError(t *testing.T) {
	errBoom := errors.New("boom")
	j, _ := New([]byte(","), 1, func([]byte) error { return errBoom })
	w := &Writer{j: j}
	_, err := w.Write([]byte("x")) // maxChunks=1 triggers immediate flush
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected boom error, got %v", err)
	}
}
