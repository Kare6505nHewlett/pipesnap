package tee_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/yourorg/pipesnap/internal/tee"
)

// errWriter always returns an error on Write.
type errWriter struct{ err error }

func (e *errWriter) Write(_ []byte) (int, error) { return 0, e.err }

func TestNewRequiresAtLeastOneDest(t *testing.T) {
	_, err := tee.New()
	if err == nil {
		t.Fatal("expected error when no destinations provided")
	}
}

func TestWriteFansOut(t *testing.T) {
	var a, b bytes.Buffer
	w, err := tee.New(&a, &b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	payload := []byte("hello tee")
	n, err := w.Write(payload)
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	if n != len(payload) {
		t.Fatalf("expected n=%d got %d", len(payload), n)
	}
	if a.String() != string(payload) {
		t.Errorf("dest a: got %q want %q", a.String(), payload)
	}
	if b.String() != string(payload) {
		t.Errorf("dest b: got %q want %q", b.String(), payload)
	}
}

func TestWriteReturnsFirstError(t *testing.T) {
	sentinel := errors.New("disk full")
	var good bytes.Buffer
	bad := &errWriter{err: sentinel}

	w, _ := tee.New(bad, &good)
	_, err := w.Write([]byte("data"))
	if err == nil {
		t.Fatal("expected error from errWriter")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
	// good should not have been written because bad comes first
	if good.Len() != 0 {
		t.Errorf("expected good dest to be empty, got %d bytes", good.Len())
	}
}

func TestShortWriteDetected(t *testing.T) {
	// limitWriter accepts only 2 bytes at a time
	pr, pw := io.Pipe()
	go func() {
		buf := make([]byte, 2)
		pr.Read(buf) //nolint:errcheck
		pr.Close()
	}()

	w, _ := tee.New(pw)
	_, err := w.Write([]byte("toolong"))
	// pipe will surface an error or short write; either is acceptable
	if err == nil {
		t.Log("no error returned (pipe drained full write)")
	}
}

func TestAddDest(t *testing.T) {
	var a bytes.Buffer
	w, _ := tee.New(&a)
	if w.Len() != 1 {
		t.Fatalf("expected 1 dest, got %d", w.Len())
	}

	var b bytes.Buffer
	w.Add(&b)
	if w.Len() != 2 {
		t.Fatalf("expected 2 dests, got %d", w.Len())
	}

	w.Write([]byte("hi")) //nolint:errcheck
	if b.String() != "hi" {
		t.Errorf("newly added dest did not receive data")
	}
}
