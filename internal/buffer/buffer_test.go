package buffer

import (
	"bytes"
	"errors"
	"testing"
)

func TestNewRejectsZeroCap(t *testing.T) {
	_, err := New(0, &bytes.Buffer{})
	if !errors.Is(err, ErrCapacity) {
		t.Fatalf("expected ErrCapacity, got %v", err)
	}
}

func TestNewRejectsNilDst(t *testing.T) {
	_, err := New(4, nil)
	if err == nil {
		t.Fatal("expected error for nil dst")
	}
}

func TestWriteBuffersUntilCapacity(t *testing.T) {
	var dst bytes.Buffer
	b, _ := New(3, &dst)

	b.Write([]byte("a"))
	b.Write([]byte("b"))
	if dst.Len() != 0 {
		t.Fatalf("expected no flush yet, got %d bytes in dst", dst.Len())
	}
	b.Write([]byte("c")) // triggers auto-flush
	if dst.String() != "abc" {
		t.Fatalf("expected 'abc', got %q", dst.String())
	}
	if b.Len() != 0 {
		t.Fatalf("buffer should be empty after flush, got %d", b.Len())
	}
}

func TestFlushEmitsPartial(t *testing.T) {
	var dst bytes.Buffer
	b, _ := New(10, &dst)

	b.Write([]byte("hello"))
	b.Write([]byte(" world"))
	if dst.Len() != 0 {
		t.Fatal("expected nothing flushed yet")
	}
	if err := b.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}
	if dst.String() != "hello world" {
		t.Fatalf("unexpected dst content: %q", dst.String())
	}
}

func TestBytesTracksSize(t *testing.T) {
	var dst bytes.Buffer
	b, _ := New(10, &dst)

	b.Write([]byte("foo"))
	b.Write([]byte("bar"))
	if b.Bytes() != 6 {
		t.Fatalf("expected 6 bytes, got %d", b.Bytes())
	}
	b.Flush()
	if b.Bytes() != 0 {
		t.Fatalf("expected 0 bytes after flush, got %d", b.Bytes())
	}
}

func TestEmptyWriteIsNoop(t *testing.T) {
	var dst bytes.Buffer
	b, _ := New(2, &dst)

	n, err := b.Write([]byte{})
	if n != 0 || err != nil {
		t.Fatalf("unexpected result for empty write: n=%d err=%v", n, err)
	}
	if b.Len() != 0 {
		t.Fatalf("expected empty buffer, got %d chunks", b.Len())
	}
}

type failWriter struct{}

func (failWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestFlushPropagatesError(t *testing.T) {
	b, _ := New(1, failWriter{})
	_, err := b.Write([]byte("x")) // auto-flush triggered
	if err == nil {
		t.Fatal("expected error from failing dst")
	}
}
