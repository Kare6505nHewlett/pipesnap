package split_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/pipesnap/internal/split"
)

func TestNewWriterRequiresLimit(t *testing.T) {
	_, err := split.NewWriter(split.Options{OutDir: t.TempDir()})
	if err == nil {
		t.Fatal("expected error with no limits")
	}
}

func TestWriterSplitsByChunks(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}

	w, err := split.NewWriter(split.Options{MaxChunks: 2, OutDir: out, Prefix: "p-"})
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	payloads := [][]byte{
		[]byte("one"), []byte("two"), []byte("three"), []byte("four"),
	}
	for _, p := range payloads {
		if _, err := w.Write(p); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if len(w.Paths) != 2 {
		t.Fatalf("expected 2 part files, got %d", len(w.Paths))
	}
	var total int
	for _, p := range w.Paths {
		total += len(readAllChunks(t, p))
	}
	if total != len(payloads) {
		t.Fatalf("expected %d chunks total, got %d", len(payloads), total)
	}
}

func TestWriterSplitsByBytes(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}

	w, err := split.NewWriter(split.Options{MaxBytes: 6, OutDir: out, Prefix: "b-"})
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	payloads := [][]byte{[]byte("abcde"), []byte("fghij"), []byte("klmno")}
	for _, p := range payloads {
		if _, err := w.Write(p); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if len(w.Paths) < 2 {
		t.Fatalf("expected at least 2 parts, got %d", len(w.Paths))
	}
}

func TestWriterCloseIdempotent(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out")
	_ = os.MkdirAll(out, 0o755)
	w, err := split.NewWriter(split.Options{MaxChunks: 1, OutDir: out})
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	_ = w.Close()
	if err := w.Close(); err != nil {
		t.Fatalf("second Close should be a no-op, got: %v", err)
	}
}
