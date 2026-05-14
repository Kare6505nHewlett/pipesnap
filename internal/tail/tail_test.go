package tail_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/pipesnap/internal/snapshot"
	"github.com/user/pipesnap/internal/tail"
)

func writeSnap(t *testing.T, chunks [][]byte) string {
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

func TestReadAllChunks(t *testing.T) {
	chunks := [][]byte{[]byte("alpha"), []byte("beta"), []byte("gamma")}
	path := writeSnap(t, chunks)

	got, err := tail.Read(path, tail.Options{N: 0})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(got))
	}
	if !bytes.Equal(got[0].Data, []byte("alpha")) {
		t.Errorf("chunk 0: got %q", got[0].Data)
	}
}

func TestReadLastN(t *testing.T) {
	chunks := [][]byte{[]byte("one"), []byte("two"), []byte("three"), []byte("four")}
	path := writeSnap(t, chunks)

	got, err := tail.Read(path, tail.Options{N: 2})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(got))
	}
	if !bytes.Equal(got[0].Data, []byte("three")) {
		t.Errorf("expected 'three', got %q", got[0].Data)
	}
	if !bytes.Equal(got[1].Data, []byte("four")) {
		t.Errorf("expected 'four', got %q", got[1].Data)
	}
}

func TestReadNLargerThanChunks(t *testing.T) {
	chunks := [][]byte{[]byte("x"), []byte("y")}
	path := writeSnap(t, chunks)

	got, err := tail.Read(path, tail.Options{N: 100})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(got))
	}
}

func TestReadMissingFile(t *testing.T) {
	_, err := tail.Read("/nonexistent/path.snap", tail.Options{N: 1})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestWrite(t *testing.T) {
	chunks := []tail.Chunk{
		{Index: 0, Data: []byte("hello ")},
		{Index: 1, Data: []byte("world")},
	}
	var buf bytes.Buffer
	if err := tail.Write(&buf, chunks); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if buf.String() != "hello world" {
		t.Errorf("got %q", buf.String())
	}
}

func TestWriteToFailingWriter(t *testing.T) {
	chunks := []tail.Chunk{{Index: 0, Data: []byte("data")}}
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Skip("cannot open /dev/null")
	}
	defer f.Close()
	// writing to a read-only file should fail
	if err := tail.Write(f, chunks); err == nil {
		t.Error("expected write error")
	}
}
