package annotate

import (
	"encoding/json"
	"testing"
)

func TestNewRejectsEmpty(t *testing.T) {
	_, err := New([]Annotation{})
	if err == nil {
		t.Fatal("expected error for empty annotations")
	}
}

func TestNewRejectsEmptyKey(t *testing.T) {
	_, err := New([]Annotation{{Key: "", Value: "v"}})
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestApplyInjectsField(t *testing.T) {
	a, err := New([]Annotation{{Key: "env", Value: "prod"}})
	if err != nil {
		t.Fatal(err)
	}
	out, err := a.Apply([]byte(`{"msg":"hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err := json.Unmarshal(out, &obj); err != nil {
		t.Fatal(err)
	}
	if obj["env"] != "prod" {
		t.Fatalf("expected env=prod, got %v", obj["env"])
	}
	if obj["msg"] != "hello" {
		t.Fatalf("expected msg=hello, got %v", obj["msg"])
	}
}

func TestApplyMultipleAnnotations(t *testing.T) {
	a, _ := New([]Annotation{
		{Key: "source", Value: "pipe"},
		{Key: "version", Value: 2},
	})
	out, err := a.Apply([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	json.Unmarshal(out, &obj)
	if obj["source"] != "pipe" {
		t.Fatalf("missing source annotation")
	}
	if obj["version"] == nil {
		t.Fatalf("missing version annotation")
	}
}

func TestApplyNonJSONReturnsError(t *testing.T) {
	a, _ := New([]Annotation{{Key: "k", Value: "v"}})
	_, err := a.Apply([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for non-JSON input")
	}
}

func TestApplyOverwritesExistingKey(t *testing.T) {
	a, _ := New([]Annotation{{Key: "env", Value: "staging"}})
	out, err := a.Apply([]byte(`{"env":"prod"}`))
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	json.Unmarshal(out, &obj)
	if obj["env"] != "staging" {
		t.Fatalf("expected env overwritten to staging, got %v", obj["env"])
	}
}

func TestLen(t *testing.T) {
	a, _ := New([]Annotation{{Key: "a", Value: 1}, {Key: "b", Value: 2}})
	if a.Len() != 2 {
		t.Fatalf("expected Len=2, got %d", a.Len())
	}
}
