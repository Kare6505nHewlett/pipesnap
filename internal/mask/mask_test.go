package mask

import (
	"bytes"
	"testing"
)

func TestNewRejectsNoPatterns(t *testing.T) {
	_, err := New("***")
	if err == nil {
		t.Fatal("expected error for zero patterns")
	}
}

func TestNewRejectsInvalidPattern(t *testing.T) {
	_, err := New("***", "[invalid")
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestApplySinglePattern(t *testing.T) {
	m, err := New("[REDACTED]", `password=\S+`)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	input := []byte(`user=alice password=secret123 host=db`)
	out := m.Apply(input)
	if bytes.Contains(out, []byte("secret123")) {
		t.Errorf("sensitive value not masked: %s", out)
	}
	if !bytes.Contains(out, []byte("[REDACTED]")) {
		t.Errorf("placeholder missing in output: %s", out)
	}
}

func TestApplyMultiplePatterns(t *testing.T) {
	m, err := New("***", `token=\S+`, `apikey=\S+`)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	input := []byte(`token=abc123 apikey=xyz789 name=pipesnap`)
	out := m.Apply(input)
	if bytes.Contains(out, []byte("abc123")) || bytes.Contains(out, []byte("xyz789")) {
		t.Errorf("not all sensitive values masked: %s", out)
	}
	if !bytes.Contains(out, []byte("name=pipesnap")) {
		t.Errorf("non-sensitive field was unexpectedly removed: %s", out)
	}
}

func TestApplyNoMatch(t *testing.T) {
	m, err := New("***", `secret=\S+`)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	input := []byte(`user=alice host=db`)
	out := m.Apply(input)
	if !bytes.Equal(out, input) {
		t.Errorf("expected unchanged output, got %s", out)
	}
}

func TestApplyDoesNotMutateInput(t *testing.T) {
	m, err := New("***", `pw=\S+`)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	original := []byte(`pw=hunter2`)
	copy := append([]byte{}, original...)
	m.Apply(original)
	if !bytes.Equal(original, copy) {
		t.Errorf("Apply mutated the input slice")
	}
}

func TestWriterMasksOnWrite(t *testing.T) {
	m, err := New("[X]", `\d+`)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	var buf bytes.Buffer
	w := NewWriter(&buf, m)
	n, err := w.Write([]byte("order=42 qty=7"))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len("order=42 qty=7") {
		t.Errorf("expected n=%d got %d", len("order=42 qty=7"), n)
	}
	if bytes.Contains(buf.Bytes(), []byte("42")) || bytes.Contains(buf.Bytes(), []byte("7")) {
		t.Errorf("digits not masked in writer output: %s", buf.Bytes())
	}
}
