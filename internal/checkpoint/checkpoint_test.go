package checkpoint_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"pipesnap/internal/checkpoint"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "checkpoint.json")
}

func TestSaveAndLoad(t *testing.T) {
	path := tempPath(t)
	now := time.Now().Truncate(time.Second)

	s := checkpoint.State{
		SnapshotFile:  "/tmp/snap.bin",
		BytesWritten:  1024,
		ChunksWritten: 8,
		CreatedAt:     now,
	}

	if err := checkpoint.Save(path, s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := checkpoint.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.SnapshotFile != s.SnapshotFile {
		t.Errorf("SnapshotFile: got %q, want %q", loaded.SnapshotFile, s.SnapshotFile)
	}
	if loaded.BytesWritten != s.BytesWritten {
		t.Errorf("BytesWritten: got %d, want %d", loaded.BytesWritten, s.BytesWritten)
	}
	if loaded.ChunksWritten != s.ChunksWritten {
		t.Errorf("ChunksWritten: got %d, want %d", loaded.ChunksWritten, s.ChunksWritten)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := checkpoint.Load("/nonexistent/path/checkpoint.json")
	if err != checkpoint.ErrNoCheckpoint {
		t.Errorf("expected ErrNoCheckpoint, got %v", err)
	}
}

func TestRemoveExisting(t *testing.T) {
	path := tempPath(t)
	if err := checkpoint.Save(path, checkpoint.State{}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := checkpoint.Remove(path); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected file to be removed")
	}
}

func TestRemoveMissingFile(t *testing.T) {
	if err := checkpoint.Remove("/nonexistent/checkpoint.json"); err != nil {
		t.Errorf("expected nil for missing file, got %v", err)
	}
}

func TestSaveUpdatesTimestamp(t *testing.T) {
	path := tempPath(t)
	before := time.Now()
	if err := checkpoint.Save(path, checkpoint.State{}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := checkpoint.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.UpdatedAt.Before(before) {
		t.Error("UpdatedAt should be set to current time on Save")
	}
}
