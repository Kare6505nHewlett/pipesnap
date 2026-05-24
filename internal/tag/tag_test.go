package tag

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestNewRejectsEmptyField(t *testing.T) {
	_, err := New("", "v1")
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNewAcceptsEmptyValue(t *testing.T) {
	_, err := New("env", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyInjectsField(t *testing.T) {
	tg, _ := New("env", "production")
	out, err := tg.Apply([]byte(`{"msg":"hello"}`))
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(out, &obj); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if obj["env"] != "production" {
		t.Errorf("expected env=production, got %v", obj["env"])
	}
	if obj["msg"] != "hello" {
		t.Errorf("original field lost: %v", obj["msg"])
	}
}

func TestApplyOverwritesExistingField(t *testing.T) {
	tg, _ := New("env", "staging")
	out, err := tg.Apply([]byte(`{"env":"production"}`))
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	var obj map[string]interface{}
	json.Unmarshal(out, &obj)
	if obj["env"] != "staging" {
		t.Errorf("expected env=staging, got %v", obj["env"])
	}
}

func TestApplyNonJSONPassthrough(t *testing.T) {
	tg, _ := New("env", "prod")
	input := []byte("not json at all")
	out, err := tg.Apply(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(out, input) {
		t.Errorf("expected passthrough, got %q", out)
	}
}

func TestNewWriterRejectsEmptyField(t *testing.T) {
	_, err := NewWriter(&bytes.Buffer{}, "", "v1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWriterInjectsTag(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "service", "pipesnap")
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	n, err := w.Write([]byte(`{"level":"info"}`))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n == 0 {
		t.Fatal("expected non-zero n")
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &obj); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if obj["service"] != "pipesnap" {
		t.Errorf("tag missing, got %v", obj["service"])
	}
}

func TestWriterReturnOriginalLen(t *testing.T) {
	var buf bytes.Buffer
	w, _ := NewWriter(&buf, "x", "y")
	input := []byte(`{"a":1}`)
	n, err := w.Write(input)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(input) {
		t.Errorf("expected %d, got %d", len(input), n)
	}
}
