package index_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/pipesnap/internal/index"
)

func tempPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "test.snap.idx")
}

func TestAddAndLookup(t *testing.T) {
	idx := &index.Index{}
	idx.Add(0, 0, 128)
	idx.Add(1, 256, 64)
	idx.Add(2, 512, 200)

	e, err := idx.Lookup(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Offset != 256 || e.Size != 64 {
		t.Fatalf("got offset=%d size=%d, want 256/64", e.Offset, e.Size)
	}
}

func TestLookupMissing(t *testing.T) {
	idx := &index.Index{}
	idx.Add(0, 0, 10)

	_, err := idx.Lookup(99)
	if err == nil {
		t.Fatal("expected error for missing seq")
	}
}

func TestLen(t *testing.T) {
	idx := &index.Index{}
	if idx.Len() != 0 {
		t.Fatalf("expected 0, got %d", idx.Len())
	}
	idx.Add(0, 0, 8)
	idx.Add(1, 8, 8)
	if idx.Len() != 2 {
		t.Fatalf("expected 2, got %d", idx.Len())
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := tempPath(t)
	idx := &index.Index{}
	idx.Add(0, 0, 100)
	idx.Add(1, 100, 200)

	if err := index.Save(path, idx); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := index.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", loaded.Len())
	}
	e, _ := loaded.Lookup(1)
	if e.Offset != 100 || e.Size != 200 {
		t.Fatalf("unexpected entry: %+v", e)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := index.Load("/nonexistent/path.idx")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPathFor(t *testing.T) {
	got := index.PathFor("/tmp/foo.snap")
	want := "/tmp/foo.snap.idx"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSaveCreatesFile(t *testing.T) {
	path := tempPath(t)
	idx := &index.Index{}
	idx.Add(0, 0, 32)

	if err := index.Save(path, idx); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}
