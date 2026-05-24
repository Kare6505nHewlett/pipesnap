package prefix

import (
	"bytes"
	"testing"
)

func TestNewRejectsEmptyPrefix(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Fatal("expected error for empty prefix, got nil")
	}
}

func TestNewAcceptsValidPrefix(t *testing.T) {
	pf, err := New("[INFO] ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pf == nil {
		t.Fatal("expected non-nil Prefixer")
	}
}

func TestApplyPrependsPrefix(t *testing.T) {
	pf, _ := New(">> ")
	out, err := pf.Apply([]byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := string(out), ">> hello"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestApplyEmptyDataReturnsEmpty(t *testing.T) {
	pf, _ := New(">> ")
	out, err := pf.Apply([]byte{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty output, got %q", out)
	}
}

func TestApplyDoesNotMutateInput(t *testing.T) {
	pf, _ := New("X")
	input := []byte("data")
	copy := append([]byte{}, input...)
	pf.Apply(input) //nolint:errcheck
	if !bytes.Equal(input, copy) {
		t.Error("Apply mutated the input slice")
	}
}

func TestNewWriterRejectsEmptyPrefix(t *testing.T) {
	_, err := NewWriter(&bytes.Buffer{}, "")
	if err == nil {
		t.Fatal("expected error for empty prefix")
	}
}

func TestWriterPassthrough(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "[log] ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n, err := w.Write([]byte("message"))
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	if n != len("message") {
		t.Errorf("reported n=%d, want %d", n, len("message"))
	}
	if got, want := buf.String(), "[log] message"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriterMultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	w, _ := NewWriter(&buf, "> ")
	w.Write([]byte("first"))  //nolint:errcheck
	w.Write([]byte("second")) //nolint:errcheck
	expected := "> first> second"
	if got := buf.String(); got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestWriterReturnOriginalLen(t *testing.T) {
	var buf bytes.Buffer
	w, _ := NewWriter(&buf, "PREFIX:")
	data := []byte("payload")
	n, err := w.Write(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("n=%d, want %d", n, len(data))
	}
}
