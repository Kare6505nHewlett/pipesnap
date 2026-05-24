package batch

import (
	"strings"
	"testing"
)

func TestNewRejectsEmptyField(t *testing.T) {
	_, err := New("", 10, func(string, [][]byte) error { return nil })
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNewRejectsZeroMaxSize(t *testing.T) {
	_, err := New("key", 0, func(string, [][]byte) error { return nil })
	if err == nil {
		t.Fatal("expected error for zero maxSize")
	}
}

func TestNewRejectsNilOnEmit(t *testing.T) {
	_, err := New("key", 5, nil)
	if err == nil {
		t.Fatal("expected error for nil onEmit")
	}
}

func TestGroupsSameKey(t *testing.T) {
	var got [][]byte
	b, _ := New("type", 10, func(_ string, chunks [][]byte) error {
		got = append(got, chunks...)
		return nil
	})

	_ = b.Add([]byte(`{"type":"a","v":1}`))
	_ = b.Add([]byte(`{"type":"a","v":2}`))
	// no flush yet — same key, under maxSize
	if len(got) != 0 {
		t.Fatalf("expected 0 emitted chunks, got %d", len(got))
	}

	_ = b.Flush()
	if len(got) != 2 {
		t.Fatalf("expected 2 chunks after flush, got %d", len(got))
	}
}

func TestKeyChangeTriggersFlush(t *testing.T) {
	var emitted []string
	b, _ := New("type", 10, func(key string, chunks [][]byte) error {
		emitted = append(emitted, key)
		return nil
	})

	_ = b.Add([]byte(`{"type":"a"}`))
	_ = b.Add([]byte(`{"type":"b"}`))

	if len(emitted) != 1 || emitted[0] != "a" {
		t.Fatalf("expected emit for key 'a', got %v", emitted)
	}
}

func TestMaxSizeTriggersFlush(t *testing.T) {
	flushCount := 0
	b, _ := New("type", 2, func(_ string, _ [][]byte) error {
		flushCount++
		return nil
	})

	_ = b.Add([]byte(`{"type":"x"}`))
	_ = b.Add([]byte(`{"type":"x"}`))

	if flushCount != 1 {
		t.Fatalf("expected 1 flush at maxSize, got %d", flushCount)
	}
}

func TestAddInvalidJSONReturnsError(t *testing.T) {
	b, _ := New("type", 5, func(string, [][]byte) error { return nil })
	err := b.Add([]byte(`not-json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "batch:") {
		t.Fatalf("expected error to mention 'batch:', got: %v", err)
	}
}

func TestFlushNoOpWhenEmpty(t *testing.T) {
	called := false
	b, _ := New("type", 5, func(string, [][]byte) error {
		called = true
		return nil
	})
	if err := b.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("onEmit should not be called on empty flush")
	}
}

func TestMissingFieldKeyIsEmpty(t *testing.T) {
	var keys []string
	b, _ := New("type", 10, func(key string, _ [][]byte) error {
		keys = append(keys, key)
		return nil
	})
	_ = b.Add([]byte(`{"other":"x"}`))
	_ = b.Flush()
	if len(keys) != 1 || keys[0] != "" {
		t.Fatalf("expected empty key, got %v", keys)
	}
}
