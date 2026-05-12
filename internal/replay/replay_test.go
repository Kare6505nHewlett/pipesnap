package replay_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/pipesnap/internal/replay"
	"github.com/user/pipesnap/internal/snapshot"
)

func writeSnapshot(t *testing.T, chunks [][]byte) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.snap")

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
	return path
}

func TestReplayBasic(t *testing.T) {
	chunks := [][]byte{[]byte("hello "), []byte("world\n")}
	path := writeSnapshot(t, chunks)

	var buf bytes.Buffer
	r, err := replay.New(path, replay.Options{Dest: &buf})
	if err != nil {
		t.Fatalf("replay.New: %v", err)
	}
	defer r.Close()

	n, err := r.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := "hello world\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
	if int(n) != len(want) {
		t.Errorf("bytes written = %d, want %d", n, len(want))
	}
}

func TestReplayMissingFile(t *testing.T) {
	_, err := replay.New("/nonexistent/path.snap", replay.Options{Dest: os.Stdout})
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestReplayChunkDelay(t *testing.T) {
	chunks := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	path := writeSnapshot(t, chunks)

	var buf bytes.Buffer
	delay := 10 * time.Millisecond
	r, err := replay.New(path, replay.Options{Dest: &buf, ChunkDelay: delay})
	if err != nil {
		t.Fatalf("replay.New: %v", err)
	}
	defer r.Close()

	start := time.Now()
	if _, err := r.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}
	elapsed := time.Since(start)

	// Expect at least 2 delays (between 3 chunks)
	if elapsed < 2*delay {
		t.Errorf("elapsed %v < expected minimum %v", elapsed, 2*delay)
	}
	if buf.String() != "abc" {
		t.Errorf("got %q, want %q", buf.String(), "abc")
	}
}
