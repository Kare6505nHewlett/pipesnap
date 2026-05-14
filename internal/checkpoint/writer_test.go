package checkpoint_test

import (
	"bytes"
	"testing"

	"pipesnap/internal/checkpoint"
)

func TestWriterUpdatesState(t *testing.T) {
	var buf bytes.Buffer
	path := tempPath(t)

	w := checkpoint.NewWriter(&buf, path, checkpoint.State{
		SnapshotFile: "snap.bin",
	})

	payload := []byte("hello world")
	n, err := w.Write(payload)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(payload) {
		t.Errorf("n: got %d, want %d", n, len(payload))
	}

	s := w.State()
	if s.BytesWritten != int64(len(payload)) {
		t.Errorf("BytesWritten: got %d, want %d", s.BytesWritten, len(payload))
	}
	if s.ChunksWritten != 1 {
		t.Errorf("ChunksWritten: got %d, want 1", s.ChunksWritten)
	}
}

func TestWriterPersistsCheckpoint(t *testing.T) {
	var buf bytes.Buffer
	path := tempPath(t)

	w := checkpoint.NewWriter(&buf, path, checkpoint.State{
		SnapshotFile: "snap.bin",
	})

	_, _ = w.Write([]byte("chunk1"))
	_, _ = w.Write([]byte("chunk2"))

	loaded, err := checkpoint.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ChunksWritten != 2 {
		t.Errorf("ChunksWritten: got %d, want 2", loaded.ChunksWritten)
	}
	if loaded.BytesWritten != 12 {
		t.Errorf("BytesWritten: got %d, want 12", loaded.BytesWritten)
	}
}

func TestWriterSetsCreatedAt(t *testing.T) {
	var buf bytes.Buffer
	path := tempPath(t)

	w := checkpoint.NewWriter(&buf, path, checkpoint.State{})
	if w.State().CreatedAt.IsZero() {
		t.Error("CreatedAt should be set automatically when zero")
	}
}

func TestWriterPassesDataThrough(t *testing.T) {
	var buf bytes.Buffer
	path := tempPath(t)
	w := checkpoint.NewWriter(&buf, path, checkpoint.State{})

	_, _ = w.Write([]byte("data"))
	if buf.String() != "data" {
		t.Errorf("underlying writer got %q, want %q", buf.String(), "data")
	}
}
