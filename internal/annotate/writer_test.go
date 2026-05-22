package annotate

import (
	"bytes"
	"encoding/json"
	"testing"
)

func makeAnnotator(t *testing.T) *Annotator {
	t.Helper()
	a, err := New([]Annotation{{Key: "tag", Value: "test"}})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestNewWriterRejectsNilDst(t *testing.T) {
	_, err := NewWriter(nil, makeAnnotator(t))
	if err == nil {
		t.Fatal("expected error for nil dst")
	}
}

func TestNewWriterRejectsNilAnnotator(t *testing.T) {
	_, err := NewWriter(&bytes.Buffer{}, nil)
	if err == nil {
		t.Fatal("expected error for nil annotator")
	}
}

func TestWriterPassthrough(t *testing.T) {
	buf := &bytes.Buffer{}
	w, err := NewWriter(buf, makeAnnotator(t))
	if err != nil {
		t.Fatal(err)
	}
	n, err := w.Write([]byte(`{"x":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if n != len(`{"x":1}`) {
		t.Fatalf("expected n=%d, got %d", len(`{"x":1}`), n)
	}
	var obj map[string]any
	if err := json.Unmarshal(buf.Bytes(), &obj); err != nil {
		t.Fatal(err)
	}
	if obj["tag"] != "test" {
		t.Fatalf("expected tag=test in output")
	}
}

func TestWriterDropsInvalidChunk(t *testing.T) {
	buf := &bytes.Buffer{}
	w, _ := NewWriter(buf, makeAnnotator(t))
	_, err := w.Write([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON chunk")
	}
	if w.Drops() != 1 {
		t.Fatalf("expected 1 drop, got %d", w.Drops())
	}
	if buf.Len() != 0 {
		t.Fatal("expected nothing written to dst on error")
	}
}

func TestWriterMultipleWrites(t *testing.T) {
	buf := &bytes.Buffer{}
	w, _ := NewWriter(buf, makeAnnotator(t))
	for i := 0; i < 3; i++ {
		if _, err := w.Write([]byte(`{"i":1}`)); err != nil {
			t.Fatalf("write %d failed: %v", i, err)
		}
	}
	if w.Drops() != 0 {
		t.Fatalf("expected 0 drops, got %d", w.Drops())
	}
}
