package rotate_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/pipesnap/internal/rotate"
)

func writeSnap(t *testing.T, dir, name string, size int) {
	t.Helper()
	data := make([]byte, size)
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		t.Fatalf("writeSnap: %v", err)
	}
}

func TestNextPathHasPrefix(t *testing.T) {
	m := rotate.New(rotate.Config{Dir: t.TempDir(), Prefix: "snap-"})
	p := m.NextPath()
	base := filepath.Base(p)
	if len(base) == 0 || base[:5] != "snap-" {
		t.Errorf("expected prefix snap-, got %s", base)
	}
}

func TestRotateByMaxFiles(t *testing.T) {
	dir := t.TempDir()
	m := rotate.New(rotate.Config{Dir: dir, Prefix: "snap-", MaxFiles: 2})

	for i, name := range []string{"snap-001.snap", "snap-002.snap", "snap-003.snap"} {
		writeSnap(t, dir, name, 10)
		time.Sleep(time.Millisecond * time.Duration(i+1))
	}

	removed, err := m.Rotate()
	if err != nil {
		t.Fatalf("Rotate: %v", err)
	}
	if len(removed) != 1 {
		t.Fatalf("expected 1 removed, got %d", len(removed))
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Errorf("expected 2 remaining files, got %d", len(entries))
	}
}

func TestRotateByMaxBytes(t *testing.T) {
	dir := t.TempDir()
	m := rotate.New(rotate.Config{Dir: dir, Prefix: "snap-", MaxBytes: 25})

	for i, name := range []string{"snap-001.snap", "snap-002.snap", "snap-003.snap"} {
		writeSnap(t, dir, name, 10)
		time.Sleep(time.Millisecond * time.Duration(i+1))
	}

	removed, err := m.Rotate()
	if err != nil {
		t.Fatalf("Rotate: %v", err)
	}
	if len(removed) == 0 {
		t.Fatal("expected at least one file removed")
	}

	entries, _ := os.ReadDir(dir)
	var total int64
	for _, e := range entries {
		info, _ := e.Info()
		total += info.Size()
	}
	if total > 25 {
		t.Errorf("total size %d exceeds MaxBytes 25", total)
	}
}

func TestRotateNoopWhenUnderLimit(t *testing.T) {
	dir := t.TempDir()
	m := rotate.New(rotate.Config{Dir: dir, Prefix: "snap-", MaxFiles: 5})

	writeSnap(t, dir, "snap-001.snap", 10)
	writeSnap(t, dir, "snap-002.snap", 10)

	removed, err := m.Rotate()
	if err != nil {
		t.Fatalf("Rotate: %v", err)
	}
	if len(removed) != 0 {
		t.Errorf("expected no removals, got %d", len(removed))
	}
}

func TestRotateIgnoresNonSnapFiles(t *testing.T) {
	dir := t.TempDir()
	m := rotate.New(rotate.Config{Dir: dir, Prefix: "snap-", MaxFiles: 1})

	writeSnap(t, dir, "snap-001.snap", 10)
	// write a file that doesn't match prefix
	os.WriteFile(filepath.Join(dir, "other.log"), []byte("x"), 0o644)

	removed, err := m.Rotate()
	if err != nil {
		t.Fatalf("Rotate: %v", err)
	}
	if len(removed) != 0 {
		t.Errorf("expected no snap removals, got %v", removed)
	}
}
