package routing_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/your-org/pipesnap/internal/routing"
)

func matchPrefix(prefix string) func([]byte) bool {
	return func(p []byte) bool {
		return strings.HasPrefix(string(p), prefix)
	}
}

func TestNewRejectsNoRules(t *testing.T) {
	_, err := routing.New(nil)
	if err == nil {
		t.Fatal("expected error for empty rules, got nil")
	}
}

func TestNewRejectsNilMatch(t *testing.T) {
	var dst bytes.Buffer
	_, err := routing.New(nil, routing.Rule{Name: "bad", Match: nil, Dst: &dst})
	if err == nil {
		t.Fatal("expected error for nil Match, got nil")
	}
}

func TestNewRejectsNilDst(t *testing.T) {
	_, err := routing.New(nil, routing.Rule{Name: "bad", Match: matchPrefix("x"), Dst: nil})
	if err == nil {
		t.Fatal("expected error for nil Dst, got nil")
	}
}

func TestRouteMatchesFirstRule(t *testing.T) {
	var a, b bytes.Buffer
	r, err := routing.New(nil,
		routing.Rule{Name: "a", Match: matchPrefix("ERR"), Dst: &a},
		routing.Rule{Name: "b", Match: matchPrefix("INF"), Dst: &b},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := r.Write([]byte("ERR something failed")); err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if _, err := r.Write([]byte("INF all good")); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	if !strings.Contains(a.String(), "ERR") {
		t.Errorf("expected 'a' buffer to contain ERR chunk, got %q", a.String())
	}
	if !strings.Contains(b.String(), "INF") {
		t.Errorf("expected 'b' buffer to contain INF chunk, got %q", b.String())
	}
	if a.Len() > 0 && strings.Contains(a.String(), "INF") {
		t.Error("INF chunk should not appear in 'a' buffer")
	}
}

func TestRouteFallback(t *testing.T) {
	var matched, fallback bytes.Buffer
	r, err := routing.New(&fallback,
		routing.Rule{Name: "err", Match: matchPrefix("ERR"), Dst: &matched},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r.Write([]byte("DBG debug line"))

	if fallback.Len() == 0 {
		t.Error("expected fallback to receive unmatched chunk")
	}
	if matched.Len() != 0 {
		t.Error("matched buffer should be empty")
	}
}

func TestRouteDropsWhenNoFallback(t *testing.T) {
	var matched bytes.Buffer
	r, _ := routing.New(nil,
		routing.Rule{Name: "err", Match: matchPrefix("ERR"), Dst: &matched},
	)

	n, err := r.Write([]byte("DBG unmatched"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len("DBG unmatched") {
		t.Errorf("expected n=%d, got %d", len("DBG unmatched"), n)
	}
	if matched.Len() != 0 {
		t.Error("matched buffer should be empty when no fallback and no match")
	}
}
