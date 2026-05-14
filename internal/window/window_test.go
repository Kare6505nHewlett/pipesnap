package window

import (
	"bytes"
	"testing"
)

func TestNewDefaultCapacity(t *testing.T) {
	b := New(0)
	if b.capacity != 1 {
		t.Fatalf("expected capacity 1, got %d", b.capacity)
	}
}

func TestWriteAndSnapshotOrdered(t *testing.T) {
	b := New(4)
	b.Write([]byte("a"))
	b.Write([]byte("b"))
	b.Write([]byte("c"))

	snap := b.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(snap))
	}
	expected := []string{"a", "b", "c"}
	for i, ch := range snap {
		if !bytes.Equal(ch.Data, []byte(expected[i])) {
			t.Errorf("chunk %d: got %q, want %q", i, ch.Data, expected[i])
		}
	}
}

func TestEvictsOldestWhenFull(t *testing.T) {
	b := New(3)
	b.Write([]byte("x"))
	b.Write([]byte("y"))
	b.Write([]byte("z"))
	b.Write([]byte("w")) // evicts "x"

	snap := b.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(snap))
	}
	expected := []string{"y", "z", "w"}
	for i, ch := range snap {
		if !bytes.Equal(ch.Data, []byte(expected[i])) {
			t.Errorf("chunk %d: got %q, want %q", i, ch.Data, expected[i])
		}
	}
}

func TestSnapshotIsolation(t *testing.T) {
	b := New(2)
	b.Write([]byte("hello"))

	snap := b.Snapshot()
	snap[0].Data[0] = 'X' // mutate returned copy

	snap2 := b.Snapshot()
	if snap2[0].Data[0] != 'h' {
		t.Error("snapshot mutation affected internal buffer")
	}
}

func TestLen(t *testing.T) {
	b := New(5)
	if b.Len() != 0 {
		t.Fatal("expected 0")
	}
	b.Write([]byte("a"))
	b.Write([]byte("b"))
	if b.Len() != 2 {
		t.Fatalf("expected 2, got %d", b.Len())
	}
}

func TestReset(t *testing.T) {
	b := New(3)
	b.Write([]byte("a"))
	b.Write([]byte("b"))
	b.Reset()

	if b.Len() != 0 {
		t.Fatalf("expected 0 after reset, got %d", b.Len())
	}
	if len(b.Snapshot()) != 0 {
		t.Fatal("expected empty snapshot after reset")
	}
}

func TestOverwriteMultipleTimes(t *testing.T) {
	b := New(2)
	for i := 0; i < 10; i++ {
		b.Write([]byte{byte('0' + i)})
	}
	// Should retain last 2: '8', '9'
	snap := b.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2, got %d", len(snap))
	}
	if snap[0].Data[0] != '8' || snap[1].Data[0] != '9' {
		t.Errorf("unexpected tail: %q %q", snap[0].Data, snap[1].Data)
	}
}
