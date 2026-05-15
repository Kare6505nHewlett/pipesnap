package patch

import (
	"testing"
)

func TestNewRejectsEmptyFind(t *testing.T) {
	_, err := New([]Sub{{Find: []byte{}, Replace: []byte("x")}})
	if err == nil {
		t.Fatal("expected error for empty Find, got nil")
	}
}

func TestNewAcceptsEmptyReplace(t *testing.T) {
	_, err := New([]Sub{{Find: []byte("foo"), Replace: nil}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplySingleSub(t *testing.T) {
	p, _ := New([]Sub{{Find: []byte("foo"), Replace: []byte("bar")}})
	got := p.Apply([]byte("foo baz foo"))
	want := []byte("bar baz bar")
	if string(got) != string(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestApplyMultipleSubsOrdered(t *testing.T) {
	p, _ := New([]Sub{
		{Find: []byte("hello"), Replace: []byte("hi")},
		{Find: []byte("hi"), Replace: []byte("hey")},
	})
	got := p.Apply([]byte("hello world"))
	// second sub replaces the output of the first
	want := "hey world"
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestApplyNoMatch(t *testing.T) {
	p, _ := New([]Sub{{Find: []byte("xyz"), Replace: []byte("abc")}})
	input := []byte("nothing to replace")
	got := p.Apply(input)
	if string(got) != string(input) {
		t.Fatalf("got %q, want %q", got, input)
	}
}

func TestApplyDeletesWhenReplaceEmpty(t *testing.T) {
	p, _ := New([]Sub{{Find: []byte("bad"), Replace: nil}})
	got := p.Apply([]byte("good bad good"))
	want := "good  good"
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestLen(t *testing.T) {
	p, _ := New([]Sub{
		{Find: []byte("a"), Replace: []byte("b")},
		{Find: []byte("c"), Replace: []byte("d")},
	})
	if p.Len() != 2 {
		t.Fatalf("expected Len 2, got %d", p.Len())
	}
}
