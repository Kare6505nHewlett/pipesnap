package audit

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewRejectsNilDst(t *testing.T) {
	_, err := New(nil, "x", 0)
	if err == nil {
		t.Fatal("expected error for nil dst")
	}
}

func TestRecordWritesJSON(t *testing.T) {
	var buf bytes.Buffer
	l, err := New(&buf, "test", 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := l.Record([]byte("hello world")); err != nil {
		t.Fatalf("Record: %v", err)
	}
	var entry Entry
	if err := json.Unmarshal(bytes.TrimRight(buf.Bytes(), "\n"), &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entry.Label != "test" {
		t.Errorf("label = %q, want %q", entry.Label, "test")
	}
	if entry.Bytes != 11 {
		t.Errorf("bytes = %d, want 11", entry.Bytes)
	}
	if entry.Preview != "hello world" {
		t.Errorf("preview = %q, want %q", entry.Preview, "hello world")
	}
	if entry.Truncated {
		t.Error("truncated should be false for short chunk")
	}
}

func TestRecordTruncatesLongChunk(t *testing.T) {
	var buf bytes.Buffer
	l, _ := New(&buf, "", 8)
	data := []byte("abcdefghijklmnop") // 16 bytes
	_ = l.Record(data)
	var entry Entry
	_ = json.Unmarshal(bytes.TrimRight(buf.Bytes(), "\n"), &entry)
	if entry.Preview != "abcdefgh" {
		t.Errorf("preview = %q, want %q", entry.Preview, "abcdefgh")
	}
	if !entry.Truncated {
		t.Error("expected truncated=true")
	}
}

func TestRecordNewlineDelimited(t *testing.T) {
	var buf bytes.Buffer
	l, _ := New(&buf, "", 0)
	_ = l.Record([]byte("a"))
	_ = l.Record([]byte("b"))
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
}

func TestDefaultPreviewLen(t *testing.T) {
	var buf bytes.Buffer
	l, _ := New(&buf, "", 0) // 0 → default 64
	data := bytes.Repeat([]byte("x"), 100)
	_ = l.Record(data)
	var entry Entry
	_ = json.Unmarshal(bytes.TrimRight(buf.Bytes(), "\n"), &entry)
	if len(entry.Preview) != 64 {
		t.Errorf("preview len = %d, want 64", len(entry.Preview))
	}
}
