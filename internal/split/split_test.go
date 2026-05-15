package split_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/pipesnap/internal/snapshot"
	"github.com/yourorg/pipesnap/internal/split"
)

func writeSnap(t *testing.T, path string, chunks [][]byte) {
	t.Helper()
	w, err := snapshot.NewWriter(path)
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
}

func readAllChunks(t *testing.T, path string) [][]byte {
	t.Helper()
	r, err := snapshot.OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer r.Close()
	var out [][]byte
	for {
		c, err := r.Next()
		if err != nil {
			break
		}
		out = append(out, c)
	}
	return out
}

func TestSplitNoOptsError(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.snap")
	writeSnap(t, src, [][]byte{[]byte("hello")})
	_, err := split.File(src, split.Options{})
	if err == nil {
		t.Fatal("expected error when no limits set")
	}
}

func TestSplitByMaxChunks(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.snap")
	chunks := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")}
	writeSnap(t, src, chunks)

	out := filepath.Join(dir, "parts")
	paths, err := split.File(src, split.Options{MaxChunks: 2, OutDir: out, Prefix: "part-"})
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if len(paths) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(paths))
	}
	var all [][]byte
	for _, p := range paths {
		all = append(all, readAllChunks(t, p)...)
	}
	if len(all) != len(chunks) {
		t.Fatalf("expected %d total chunks, got %d", len(chunks), len(all))
	}
}

func TestSplitByMaxBytes(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.snap")
	chunks := [][]byte{[]byte("aaaa"), []byte("bbbb"), []byte("cccc")}
	writeSnap(t, src, chunks)

	out := filepath.Join(dir, "parts")
	paths, err := split.File(src, split.Options{MaxBytes: 5, OutDir: out, Prefix: "b-"})
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if len(paths) < 2 {
		t.Fatalf("expected at least 2 parts, got %d", len(paths))
	}
	var all [][]byte
	for _, p := range paths {
		all = append(all, readAllChunks(t, p)...)
	}
	if len(all) != len(chunks) {
		t.Fatalf("chunk count mismatch: want %d got %d", len(chunks), len(all))
	}
}

func TestSplitMissingSource(t *testing.T) {
	_, err := split.File("/nonexistent/file.snap", split.Options{MaxChunks: 1})
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestSplitCreatesOutDir(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.snap")
	writeSnap(t, src, [][]byte{[]byte("x")})
	out := filepath.Join(dir, "new", "sub")
	_, err := split.File(src, split.Options{MaxChunks: 1, OutDir: out})
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if _, serr := os.Stat(out); serr != nil {
		t.Fatalf("output dir not created: %v", serr)
	}
}
