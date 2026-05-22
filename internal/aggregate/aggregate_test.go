package aggregate

import (
	"bytes"
	"testing"
)

func TestNewRejectsZeroSize(t *testing.T) {
	_, err := New(0, func([]byte) error { return nil })
	if err == nil {
		t.Fatal("expected error for size=0")
	}
}

func TestNewRejectsNilOnEmit(t *testing.T) {
	_, err := New(2, nil)
	if err == nil {
		t.Fatal("expected error for nil onEmit")
	}
}

func TestAddBatchesChunks(t *testing.T) {
	var got [][]byte
	a, _ := New(3, func(p []byte) error {
		copy := make([]byte, len(p))
		copy2 := copy
		_ = copy2
		got = append(got, append([]byte(nil), p...))
		return nil
	})

	for _, s := range []string{"aa", "bb", "cc"} {
		if err := a.Add([]byte(s)); err != nil {
			t.Fatalf("Add: %v", err)
		}
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 emission, got %d", len(got))
	}
	if !bytes.Equal(got[0], []byte("aabbcc")) {
		t.Errorf("unexpected payload: %q", got[0])
	}
}

func TestFlushEmitsPartialBatch(t *testing.T) {
	var got [][]byte
	a, _ := New(5, func(p []byte) error {
		got = append(got, append([]byte(nil), p...))
		return nil
	})

	_ = a.Add([]byte("hello"))
	_ = a.Add([]byte(" world"))

	if a.Len() != 2 {
		t.Fatalf("expected Len=2, got %d", a.Len())
	}

	if err := a.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 emission after flush, got %d", len(got))
	}
	if !bytes.Equal(got[0], []byte("hello world")) {
		t.Errorf("unexpected payload: %q", got[0])
	}
	if a.Len() != 0 {
		t.Errorf("expected empty buffer after flush, got %d", a.Len())
	}
}

func TestFlushNoOpWhenEmpty(t *testing.T) {
	called := false
	a, _ := New(2, func([]byte) error {
		called = true
		return nil
	})
	if err := a.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("onEmit should not be called on empty flush")
	}
}

func TestMultipleBatches(t *testing.T) {
	count := 0
	a, _ := New(2, func([]byte) error {
		count++
		return nil
	})
	for i := 0; i < 6; i++ {
		_ = a.Add([]byte("x"))
	}
	if count != 3 {
		t.Errorf("expected 3 emissions, got %d", count)
	}
}
