package truncate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/pipesnap/internal/snapshot"
	"github.com/user/pipesnap/internal/truncate"
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

func readSnap(t *testing.T, path string) [][]byte {
	t.Helper()
	r, err := snapshot.OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer r.Close()
	var out [][]byte
	for {
		b, err := r.ReadChunk()
		if err != nil {
			break
		}
		out = append(out, b)
	}
	return out
}

func TestTruncateNoOp(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "snap.bin")
	chunks := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	writeSnap(t, p, chunks)

	if err := truncate.File(p, p, truncate.Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := readSnap(t, p)
	if len(got) != 3 {
		t.Fatalf("want 3 chunks, got %d", len(got))
	}
}

func TestTruncateByMaxChunks(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "snap.bin")
	chunks := [][]byte{[]byte("first"), []byte("second"), []byte("third"), []byte("fourth")}
	writeSnap(t, p, chunks)

	if err := truncate.File(p, p, truncate.Options{MaxChunks: 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := readSnap(t, p)
	if len(got) != 2 {
		t.Fatalf("want 2 chunks, got %d", len(got))
	}
	if string(got[0]) != "third" || string(got[1]) != "fourth" {
		t.Fatalf("unexpected chunks: %v", got)
	}
}

func TestTruncateByMaxBytes(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "snap.bin")
	chunks := [][]byte{[]byte("aaaa"), []byte("bbbb"), []byte("cccc")}
	writeSnap(t, p, chunks)

	// keep at most 8 bytes → drops first chunk
	if err := truncate.File(p, p, truncate.Options{MaxBytes: 8}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := readSnap(t, p)
	if len(got) != 2 {
		t.Fatalf("want 2 chunks, got %d", len(got))
	}
}

func TestTruncateMissingSource(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "missing.bin")
	err := truncate.File(p, p, truncate.Options{MaxChunks: 1})
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestTruncateDifferentDstPath(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	dst := filepath.Join(dir, "dst.bin")
	chunks := [][]byte{[]byte("x"), []byte("y"), []byte("z")}
	writeSnap(t, src, chunks)

	if err := truncate.File(src, dst, truncate.Options{MaxChunks: 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatal("source should still exist")
	}
	got := readSnap(t, dst)
	if len(got) != 2 {
		t.Fatalf("want 2 chunks in dst, got %d", len(got))
	}
}
