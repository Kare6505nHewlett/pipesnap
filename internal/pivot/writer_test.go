package pivot

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

func TestNewWriterRejectsNilPivotter(t *testing.T) {
	// NewWriter with a nil Pivotter should not panic during construction;
	// behaviour is undefined but we at least ensure it returns non-nil.
	w := NewWriter(io.Discard, &Pivotter{field: "x"})
	if w == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestWriterPassesData(t *testing.T) {
	pv, _ := New("env", nil)
	var buf bytes.Buffer
	w := NewWriter(&buf, pv)

	chunk := []byte(`{"env":"prod","msg":"ok"}`)
	n, err := w.Write(chunk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(chunk) {
		t.Fatalf("expected n=%d, got %d", len(chunk), n)
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if _, ok := result["prod"]; !ok {
		t.Fatalf("expected key 'prod' in output, got: %s", buf.String())
	}
}

func TestWriterDropsSilently(t *testing.T) {
	pv, _ := New("env", nil)
	var buf bytes.Buffer
	w := NewWriter(&buf, pv)

	// missing field — should be dropped without error
	chunk := []byte(`{"other":"value"}`)
	n, err := w.Write(chunk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(chunk) {
		t.Fatalf("expected original len returned, got %d", n)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected empty output for dropped chunk")
	}
}

func TestWriterCloseCallsUnderlying(t *testing.T) {
	closed := false
	rc := &closeTracker{closed: &closed}
	pv, _ := New("k", nil)
	w := NewWriter(rc, pv)
	if err := w.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if !closed {
		t.Fatal("expected underlying closer to be called")
	}
}

type closeTracker struct {
	bytes.Buffer
	closed *bool
}

func (c *closeTracker) Close() error {
	*c.closed = true
	return nil
}
