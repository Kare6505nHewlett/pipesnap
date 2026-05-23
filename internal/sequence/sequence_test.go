package sequence_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/user/pipesnap/internal/sequence"
)

func TestNewRejectsEmptyField(t *testing.T) {
	_, err := sequence.New("", 0)
	if err == nil {
		t.Fatal("expected error for empty field name")
	}
}

func TestNewAcceptsValidField(t *testing.T) {
	s, err := sequence.New("seq", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil Sequencer")
	}
}

func TestApplyInjectsField(t *testing.T) {
	s, _ := sequence.New("_seq", 0)
	out, err := s.Apply([]byte(`{"msg":"hello"}`))
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(out, &obj); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if _, ok := obj["_seq"]; !ok {
		t.Error("expected _seq field in output")
	}
}

func TestApplyMonotonicallyIncreasing(t *testing.T) {
	s, _ := sequence.New("n", 0)
	var prev float64 = -1
	for i := 0; i < 5; i++ {
		out, err := s.Apply([]byte(`{}`))
		if err != nil {
			t.Fatalf("Apply error: %v", err)
		}
		var obj map[string]interface{}
		json.Unmarshal(out, &obj)
		cur := obj["n"].(float64)
		if cur <= prev {
			t.Errorf("sequence not increasing: got %v after %v", cur, prev)
		}
		prev = cur
	}
}

func TestApplyNonJSONReturnsError(t *testing.T) {
	s, _ := sequence.New("seq", 0)
	_, err := s.Apply([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for non-JSON input")
	}
}

func TestApplyStartsAtGivenOffset(t *testing.T) {
	s, _ := sequence.New("seq", 100)
	out, _ := s.Apply([]byte(`{}`))
	var obj map[string]interface{}
	json.Unmarshal(out, &obj)
	if obj["seq"].(float64) != 100 {
		t.Errorf("expected start=100, got %v", obj["seq"])
	}
}

func TestWriterPassthrough(t *testing.T) {
	s, _ := sequence.New("seq", 0)
	var buf bytes.Buffer
	w := sequence.NewWriter(&buf, s)
	n, err := w.Write([]byte(`{"x":1}`))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != len(`{"x":1}`) {
		t.Errorf("expected n=%d, got %d", len(`{"x":1}`), n)
	}
	if buf.Len() == 0 {
		t.Error("expected output in buffer")
	}
}

func TestWriterDropsInvalidChunk(t *testing.T) {
	s, _ := sequence.New("seq", 0)
	var buf bytes.Buffer
	w := sequence.NewWriter(&buf, s)
	n, err := w.Write([]byte(`not json`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len("not json") {
		t.Errorf("expected original len returned")
	}
	if w.Drops != 1 {
		t.Errorf("expected 1 drop, got %d", w.Drops)
	}
	if buf.Len() != 0 {
		t.Error("expected nothing written to dst")
	}
}
