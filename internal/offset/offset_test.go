package offset_test

import (
	"os"
	"testing"

	"github.com/user/pipesnap/internal/offset"
	"github.com/user/pipesnap/internal/snapshot"
)

func writeSnap(t *testing.T, chunks [][]byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "snap-*.bin")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	path := f.Name()
	f.Close()

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

func TestBuildReturnsOneEntryPerChunk(t *testing.T) {
	chunks := [][]byte{[]byte("alpha"), []byte("beta"), []byte("gamma")}
	path := writeSnap(t, chunks)

	entries, err := offset.Build(path)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(entries) != len(chunks) {
		t.Fatalf("expected %d entries, got %d", len(chunks), len(entries))
	}
	for i, e := range entries {
		if e.ChunkIndex != int64(i) {
			t.Errorf("entry[%d].ChunkIndex = %d, want %d", i, e.ChunkIndex, i)
		}
	}
}

func TestBuildOffsetsAreStrictlyIncreasing(t *testing.T) {
	chunks := [][]byte{[]byte("x"), []byte("yy"), []byte("zzz")}
	path := writeSnap(t, chunks)

	entries, err := offset.Build(path)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	for i := 1; i < len(entries); i++ {
		if entries[i].ByteOffset <= entries[i-1].ByteOffset {
			t.Errorf("offsets not increasing: [%d]=%d <= [%d]=%d",
				i, entries[i].ByteOffset, i-1, entries[i-1].ByteOffset)
		}
	}
}

func TestBuildMissingFile(t *testing.T) {
	_, err := offset.Build("/nonexistent/snap.bin")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestBuildEmptySnapshot(t *testing.T) {
	path := writeSnap(t, nil)

	entries, err := offset.Build(path)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries for empty snapshot, got %d", len(entries))
	}
}

func TestSeekToMovesPosition(t *testing.T) {
	chunks := [][]byte{[]byte("first"), []byte("second")}
	path := writeSnap(t, chunks)

	entries, err := offset.Build(path)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(entries) < 2 {
		t.Skip("need at least 2 entries")
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	if err := offset.SeekTo(f, entries[1].ByteOffset); err != nil {
		t.Fatalf("SeekTo: %v", err)
	}
	pos, _ := f.Seek(0, 1)
	if pos != entries[1].ByteOffset {
		t.Errorf("position = %d, want %d", pos, entries[1].ByteOffset)
	}
}
