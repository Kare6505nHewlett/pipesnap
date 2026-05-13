package transform_test

import (
	"testing"

	"github.com/user/pipesnap/internal/transform"
)

func TestUpperCase(t *testing.T) {
	out, keep := transform.UpperCase([]byte("hello world"))
	if !keep {
		t.Fatal("expected keep=true")
	}
	if string(out) != "HELLO WORLD" {
		t.Fatalf("got %q", out)
	}
}

func TestLowerCase(t *testing.T) {
	out, keep := transform.LowerCase([]byte("HELLO WORLD"))
	if !keep || string(out) != "hello world" {
		t.Fatalf("unexpected result keep=%v out=%q", keep, out)
	}
}

func TestReplaceAll(t *testing.T) {
	fn := transform.ReplaceAll("foo", "bar")
	out, keep := fn([]byte("foo is foo"))
	if !keep || string(out) != "bar is bar" {
		t.Fatalf("unexpected result keep=%v out=%q", keep, out)
	}
}

func TestStripControlKeepsNormal(t *testing.T) {
	out, keep := transform.StripControl([]byte("normal text\n"))
	if !keep || string(out) != "normal text\n" {
		t.Fatalf("unexpected result keep=%v out=%q", keep, out)
	}
}

func TestStripControlRemovesEscapes(t *testing.T) {
	out, keep := transform.StripControl([]byte("text\x1b[31mred\x1b[0m"))
	if !keep {
		t.Fatal("expected keep=true")
	}
	if string(out) != "text[31mred[0m" {
		t.Fatalf("got %q", out)
	}
}

func TestStripControlDropsEmptyResult(t *testing.T) {
	_, keep := transform.StripControl([]byte("\x01\x02\x03"))
	if keep {
		t.Fatal("expected keep=false for all-control input")
	}
}

func TestTruncate(t *testing.T) {
	fn := transform.Truncate(5)
	out, keep := fn([]byte("hello world"))
	if !keep || string(out) != "hello" {
		t.Fatalf("unexpected result keep=%v out=%q", keep, out)
	}
}

func TestTruncatePreservesNewline(t *testing.T) {
	fn := transform.Truncate(5)
	out, keep := fn([]byte("hello world\n"))
	if !keep || string(out) != "hello\n" {
		t.Fatalf("unexpected result keep=%v out=%q", keep, out)
	}
}

func TestTruncateNoOpWhenShort(t *testing.T) {
	fn := transform.Truncate(100)
	out, keep := fn([]byte("hi"))
	if !keep || string(out) != "hi" {
		t.Fatalf("unexpected result keep=%v out=%q", keep, out)
	}
}

func TestChain(t *testing.T) {
	fn := transform.Chain(
		transform.UpperCase,
		transform.ReplaceAll("HELLO", "HI"),
	)
	out, keep := fn([]byte("hello world"))
	if !keep || string(out) != "HI WORLD" {
		t.Fatalf("unexpected result keep=%v out=%q", keep, out)
	}
}

func TestChainShortCircuit(t *testing.T) {
	called := false
	marker := func(b []byte) ([]byte, bool) {
		called = true
		return b, true
	}
	fn := transform.Chain(transform.StripControl, marker)
	_, keep := fn([]byte("\x01\x02"))
	if keep {
		t.Fatal("expected keep=false")
	}
	if called {
		t.Fatal("second func should not have been called")
	}
}
