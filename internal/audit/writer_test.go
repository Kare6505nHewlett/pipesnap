package audit

import (
	"bytes"
	"encoding/json"
	"testing"
)

func makeLogger(t *testing.T, buf *bytes.Buffer) *Logger {
	t.Helper()
	l, err := New(buf, "wtest", 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return l
}

func TestNewWriterRejectsNilDst(t *testing.T) {
	var logBuf bytes.Buffer
	l := makeLogger(t, &logBuf)
	_, err := NewWriter(nil, l)
	if err == nil {
		t.Fatal("expected error for nil dst")
	}
}

func TestNewWriterRejectsNilLogger(t *testing.T) {
	var dst bytes.Buffer
	_, err := NewWriter(&dst, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestWriterPassesDataThrough(t *testing.T) {
	var dst, logBuf bytes.Buffer
	l := makeLogger(t, &logBuf)
	w, err := NewWriter(&dst, l)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	payload := []byte("pipeline chunk")
	n, err := w.Write(payload)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(payload) {
		t.Errorf("n = %d, want %d", n, len(payload))
	}
	if !bytes.Equal(dst.Bytes(), payload) {
		t.Errorf("dst = %q, want %q", dst.Bytes(), payload)
	}
}

func TestWriterAuditsChunk(t *testing.T) {
	var dst, logBuf bytes.Buffer
	l := makeLogger(t, &logBuf)
	w, _ := NewWriter(&dst, l)
	_, _ = w.Write([]byte("audit me"))

	var entry Entry
	if err := json.Unmarshal(bytes.TrimRight(logBuf.Bytes(), "\n"), &entry); err != nil {
		t.Fatalf("unmarshal audit entry: %v", err)
	}
	if entry.Bytes != 8 {
		t.Errorf("entry.Bytes = %d, want 8", entry.Bytes)
	}
	if entry.Preview != "audit me" {
		t.Errorf("entry.Preview = %q, want %q", entry.Preview, "audit me")
	}
}

func TestWriterCloseCallsUnderlying(t *testing.T) {
	var logBuf bytes.Buffer
	l := makeLogger(t, &logBuf)
	closed := false
	dst := &closeTracker{closed: &closed}
	w, _ := NewWriter(dst, l)
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if !closed {
		t.Error("expected underlying Close to be called")
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
