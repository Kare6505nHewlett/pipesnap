package pivot

import (
	"encoding/json"
	"testing"
)

func TestNewRejectsEmptyField(t *testing.T) {
	_, err := New("", nil)
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNewAcceptsValidField(t *testing.T) {
	pv, err := New("type", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pv == nil {
		t.Fatal("expected non-nil Pivotter")
	}
}

func TestApplyPivotsOnField(t *testing.T) {
	pv, _ := New("kind", nil)
	chunk := []byte(`{"kind":"event","data":"hello"}`)
	out, ok := pv.Apply(chunk)
	if !ok {
		t.Fatal("expected ok=true")
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if _, exists := result["event"]; !exists {
		t.Fatalf("expected top-level key 'event', got: %s", out)
	}
}

func TestApplyDropsNonJSON(t *testing.T) {
	var dropped [][]byte
	pv, _ := New("kind", func(b []byte) { dropped = append(dropped, b) })

	_, ok := pv.Apply([]byte("not json"))
	if ok {
		t.Fatal("expected ok=false for non-JSON input")
	}
	if len(dropped) != 1 {
		t.Fatalf("expected 1 drop callback, got %d", len(dropped))
	}
}

func TestApplyDropsMissingField(t *testing.T) {
	pv, _ := New("kind", nil)
	_, ok := pv.Apply([]byte(`{"other":"value"}`))
	if ok {
		t.Fatal("expected ok=false when field is missing")
	}
	if pv.Dropped() != 1 {
		t.Fatalf("expected Dropped()=1, got %d", pv.Dropped())
	}
}

func TestApplyDropsNonStringField(t *testing.T) {
	pv, _ := New("kind", nil)
	_, ok := pv.Apply([]byte(`{"kind":42}`))
	if ok {
		t.Fatal("expected ok=false when field value is not a string")
	}
}

func TestSeenAndDroppedCounters(t *testing.T) {
	pv, _ := New("k", nil)
	pv.Apply([]byte(`{"k":"a"}`))  // ok
	pv.Apply([]byte(`{"k":"b"}`))  // ok
	pv.Apply([]byte(`{"x":"c"}`))  // drop

	if pv.Seen() != 3 {
		t.Fatalf("expected Seen()=3, got %d", pv.Seen())
	}
	if pv.Dropped() != 1 {
		t.Fatalf("expected Dropped()=1, got %d", pv.Dropped())
	}
}
