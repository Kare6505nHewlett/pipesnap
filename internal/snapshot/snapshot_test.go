package snapshot_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/pipesnap/internal/snapshot"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.snap")

	original := []byte("hello\nworld\nfoo bar baz\n")

	// Write snapshot
	w, err := snapshot.NewWriter(path, "test-label")
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	if _, err := w.Write(original); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close writer: %v", err)
	}

	// Verify file exists and is non-empty
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatal("snapshot file is empty")
	}

	// Read snapshot back
	meta, rc, err := snapshot.OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer rc.Close()

	if meta.Version != 1 {
		t.Errorf("expected version 1, got %d", meta.Version)
	}
	if meta.Label != "test-label" {
		t.Errorf("expected label 'test-label', got %q", meta.Label)
	}
	if meta.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Errorf("data mismatch:\n  got  %q\n  want %q", got, original)
	}
}

func TestOpenReaderMissingFile(t *testing.T) {
	_, _, err := snapshot.OpenReader("/nonexistent/path.snap")
	if err == nil {
		t.Fatal("expected error opening missing file, got nil")
	}
}

func TestWriterMultipleChunks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "chunks.snap")

	w, err := snapshot.NewWriter(path, "")
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	chunks := [][]byte{[]byte("chunk1\n"), []byte("chunk2\n"), []byte("chunk3\n")}
	for _, c := range chunks {
		if _, err := w.Write(c); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	w.Close()

	_, rc, err := snapshot.OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer rc.Close()

	got, _ := io.ReadAll(rc)
	want := []byte("chunk1\nchunk2\nchunk3\n")
	if !bytes.Equal(got, want) {
		t.Errorf("chunks mismatch: got %q, want %q", got, want)
	}
}
