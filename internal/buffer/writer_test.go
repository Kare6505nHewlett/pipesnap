package buffer

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

// nopCloser wraps a bytes.Buffer to satisfy io.WriteCloser.
type nopCloser struct{ *bytes.Buffer }

func (nopCloser) Close() error { return nil }

func TestNewWriterRejectsNilDst(t *testing.T) {
	_, err := NewWriter(4, nil)
	if err == nil {
		t.Fatal("expected error for nil dst")
	}
}

func TestNewWriterRejectsZeroCap(t *testing.T) {
	_, err := NewWriter(0, nopCloser{&bytes.Buffer{}})
	if !errors.Is(err, ErrCapacity) {
		t.Fatalf("expected ErrCapacity, got %v", err)
	}
}

func TestWriterBuffersAndFlushesOnClose(t *testing.T) {
	buf := &bytes.Buffer{}
	w, err := NewWriter(10, nopCloser{buf})
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	w.Write([]byte("hello"))
	w.Write([]byte(" world"))

	if buf.Len() != 0 {
		t.Fatal("expected no flush before close")
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if buf.String() != "hello world" {
		t.Fatalf("unexpected content: %q", buf.String())
	}
}

func TestWriterAutoFlushAtCapacity(t *testing.T) {
	buf := &bytes.Buffer{}
	w, _ := NewWriter(2, nopCloser{buf})

	w.Write([]byte("x"))
	w.Write([]byte("y")) // triggers flush

	if buf.String() != "xy" {
		t.Fatalf("expected 'xy' after auto-flush, got %q", buf.String())
	}
}

func TestWriterReturnOriginalLen(t *testing.T) {
	w, _ := NewWriter(4, nopCloser{&bytes.Buffer{}})
	data := []byte("some data")
	n, err := w.Write(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(data) {
		t.Fatalf("expected n=%d, got %d", len(data), n)
	}
}

// errCloser always returns an error on Close.
type errCloser struct{ io.Writer }

func (errCloser) Close() error { return errors.New("close failed") }

func TestWriterCloseReturnsError(t *testing.T) {
	w, _ := NewWriter(4, errCloser{&bytes.Buffer{}})
	if err := w.Close(); err == nil {
		t.Fatal("expected close error")
	}
}
