package reorder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func makeChunk(field string, seq int, extra string) []byte {
	m := map[string]interface{}{field: float64(seq), "data": extra}
	b, _ := json.Marshal(m)
	return b
}

func TestNewRejectsEmptyField(t *testing.T) {
	_, err := New("", 0, func([]byte) error { return nil })
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNewRejectsNilOnEmit(t *testing.T) {
	_, err := New("seq", 0, nil)
	if err == nil {
		t.Fatal("expected error for nil onEmit")
	}
}

func TestInOrderEmitsImmediately(t *testing.T) {
	var got []int
	r, _ := New("seq", 0, func(chunk []byte) error {
		var m map[string]interface{}
		json.Unmarshal(chunk, &m)
		got = append(got, int(m["seq"].(float64)))
		return nil
	})
	for i := 0; i < 5; i++ {
		if err := r.Add(makeChunk("seq", i, fmt.Sprintf("v%d", i))); err != nil {
			t.Fatalf("Add(%d): %v", i, err)
		}
	}
	if r.Buffered() != 0 {
		t.Errorf("expected 0 buffered, got %d", r.Buffered())
	}
	for i, v := range got {
		if v != i {
			t.Errorf("position %d: want %d got %d", i, i, v)
		}
	}
}

func TestOutOfOrderReorders(t *testing.T) {
	var got []int
	r, _ := New("seq", 0, func(chunk []byte) error {
		var m map[string]interface{}
		json.Unmarshal(chunk, &m)
		got = append(got, int(m["seq"].(float64)))
		return nil
	})
	order := []int{2, 0, 1, 4, 3}
	for _, i := range order {
		r.Add(makeChunk("seq", i, ""))
	}
	if r.Buffered() != 0 {
		t.Errorf("expected 0 buffered after all added, got %d", r.Buffered())
	}
	for i, v := range got {
		if v != i {
			t.Errorf("position %d: want %d got %d", i, i, v)
		}
	}
}

func TestFlushEmitsGappedChunks(t *testing.T) {
	var got []int
	r, _ := New("seq", 0, func(chunk []byte) error {
		var m map[string]interface{}
		json.Unmarshal(chunk, &m)
		got = append(got, int(m["seq"].(float64)))
		return nil
	})
	// add 0 and 2, skip 1 — 2 stays buffered
	r.Add(makeChunk("seq", 0, ""))
	r.Add(makeChunk("seq", 2, ""))
	if r.Buffered() != 1 {
		t.Errorf("expected 1 buffered, got %d", r.Buffered())
	}
	if err := r.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 emitted, got %d", len(got))
	}
}

func TestAddInvalidJSONReturnsError(t *testing.T) {
	r, _ := New("seq", 0, func([]byte) error { return nil })
	if err := r.Add([]byte("not json")); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestWriterReordersAndCloses(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "seq", 0)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	for _, i := range []int{1, 0, 2} {
		w.Write(makeChunk("seq", i, ""))
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if w.Buffered() != 0 {
		t.Errorf("expected 0 buffered after close, got %d", w.Buffered())
	}
	if buf.Len() == 0 {
		t.Error("expected data written to buffer")
	}
}
