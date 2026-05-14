package rotate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/pipesnap/internal/rotate"
)

func TestNewWriterCreatesFile(t *testing.T) {
	dir := t.TempDir()
	m := rotate.New(rotate.Config{Dir: dir, Prefix: "snap-"})

	w, err := rotate.NewWriter(m)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	defer w.Close()

	if _, err := os.Stat(w.Path()); err != nil {
		t.Errorf("expected file to exist at %s: %v", w.Path(), err)
	}
	if !strings.HasPrefix(filepath.Base(w.Path()), "snap-") {
		t.Errorf("expected snap- prefix, got %s", w.Path())
	}
}

func TestNewWriterWrite(t *testing.T) {
	dir := t.TempDir()
	m := rotate.New(rotate.Config{Dir: dir, Prefix: "snap-"})

	w, err := rotate.NewWriter(m)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	defer w.Close()

	payload := []byte("hello pipesnap")
	n, err := w.Write(payload)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(payload) {
		t.Errorf("expected %d bytes written, got %d", len(payload), n)
	}
	if w.BytesWritten() != int64(len(payload)) {
		t.Errorf("BytesWritten mismatch: got %d", w.BytesWritten())
	}
}

func TestNewWriterRotatesOldFiles(t *testing.T) {
	dir := t.TempDir()

	// Pre-populate 3 snap files.
	for _, name := range []string{"snap-001.snap", "snap-002.snap", "snap-003.snap"} {
		writeSnap(t, dir, name, 10)
	}

	m := rotate.New(rotate.Config{Dir: dir, Prefix: "snap-", MaxFiles: 2})
	w, err := rotate.NewWriter(m)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	w.Close()

	entries, _ := os.ReadDir(dir)
	// After rotation (max 2) + 1 new file = 3 total, but rotation runs before
	// the new file is created, so we should have at most MaxFiles+1 = 3.
	if len(entries) > 3 {
		t.Errorf("expected at most 3 files, got %d", len(entries))
	}
}
