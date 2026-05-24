package merge_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/pipesnap/internal/merge"
	"github.com/user/pipesnap/internal/snapshot"
)

func writeSnap(t *testing.T, dir, name string, chunks [][]byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	w, err := snapshot.NewWriter(p)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	for _, c := range chunks {
		if _, err := w.Write(c); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return p
}

func TestMergeWritesAllChunks(t *testing.T) {
	dir := t.TempDir()
	p1 := writeSnap(t, dir, "a.snap", [][]byte{[]byte("hello "), []byte("world")})
	p2 := writeSnap(t, dir, "b.snap", [][]byte{[]byte(" foo")})

	var buf bytes.Buffer
	n, err := merge.Merge(&buf, []string{p1, p2})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if want := "hello world foo"; buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
	if n != int64(len("hello world foo")) {
		t.Errorf("bytes reported: got %d, want %d", n, len("hello world foo"))
	}
}

func TestMergeNoPaths(t *testing.T) {
	var buf bytes.Buffer
	_, err := merge.Merge(&buf, nil)
	if err == nil {
		t.Fatal("expected error for empty paths")
	}
}

func TestMergeMissingFile(t *testing.T) {
	var buf bytes.Buffer
	_, err := merge.Merge(&buf, []string{"/nonexistent/path.snap"})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestCollectReturnsSortedEntries(t *testing.T) {
	dir := t.TempDir()
	p1 := writeSnap(t, dir, "b.snap", [][]byte{[]byte("beta")})
	p2 := writeSnap(t, dir, "a.snap", [][]byte{[]byte("alpha")})

	// pass in reverse order; Collect should sort
	entries, err := merge.Collect([]string{p1, p2})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if string(entries[0].Data) != "alpha" {
		t.Errorf("first entry: got %q, want %q", entries[0].Data, "alpha")
	}
	if string(entries[1].Data) != "beta" {
		t.Errorf("second entry: got %q, want %q", entries[1].Data, "beta")
	}
}

func TestCollectNoPaths(t *testing.T) {
	_, err := merge.Collect(nil)
	if err == nil {
		t.Fatal("expected error for empty paths")
	}
}

func TestCollectSourceField(t *testing.T) {
	dir := t.TempDir()
	p := writeSnap(t, dir, "x.snap", [][]byte{[]byte("data")})

	entries, err := merge.Collect([]string{p})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one entry")
	}
	if entries[0].Source != p {
		t.Errorf("Source: got %q, want %q", entries[0].Source, p)
	}
	_ = os.Remove(p)
}

func TestMergeSingleFile(t *testing.T) {
	dir := t.TempDir()
	p := writeSnap(t, dir, "only.snap", [][]byte{[]byte("sole content")})

	var buf bytes.Buffer
	n, err := merge.Merge(&buf, []string{p})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	const want = "sole content"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
	if n != int64(len(want)) {
		t.Errorf("bytes reported: got %d, want %d", n, len(want))
	}
}
