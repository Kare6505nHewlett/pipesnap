package schema

import (
	"strings"
	"testing"
)

func TestNewRejectsEmptyFields(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error for empty fields")
	}
}

func TestValidateRequiredFieldMissing(t *testing.T) {
	s, _ := New([]Field{{Name: "id", Type: TypeNumber, Required: true}})
	err := s.Validate([]byte(`{"name":"alice"}`))
	if err == nil || !strings.Contains(err.Error(), `"id"`) {
		t.Fatalf("expected missing field error, got %v", err)
	}
}

func TestValidateOptionalFieldAbsent(t *testing.T) {
	s, _ := New([]Field{{Name: "tag", Type: TypeString, Required: false}})
	if err := s.Validate([]byte(`{"other":1}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateTypeMismatch(t *testing.T) {
	s, _ := New([]Field{{Name: "count", Type: TypeNumber, Required: true}})
	err := s.Validate([]byte(`{"count":"oops"}`))
	if err == nil || !strings.Contains(err.Error(), "expected number") {
		t.Fatalf("expected type mismatch error, got %v", err)
	}
}

func TestValidateAllTypes(t *testing.T) {
	fields := []Field{
		{Name: "s", Type: TypeString, Required: true},
		{Name: "n", Type: TypeNumber, Required: true},
		{Name: "b", Type: TypeBoolean, Required: true},
		{Name: "o", Type: TypeObject, Required: true},
		{Name: "a", Type: TypeArray, Required: true},
	}
	s, _ := New(fields)
	chunk := []byte(`{"s":"hi","n":3.14,"b":true,"o":{},"a":[]}`)
	if err := s.Validate(chunk); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateInvalidJSON(t *testing.T) {
	s, _ := New([]Field{{Name: "x", Type: TypeString}})
	if err := s.Validate([]byte(`not json`)); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
