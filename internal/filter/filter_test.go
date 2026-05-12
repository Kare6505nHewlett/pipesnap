package filter_test

import (
	"testing"

	"github.com/pipesnap/pipesnap/internal/filter"
)

func TestGrepFilterMatch(t *testing.T) {
	f, err := filter.GrepFilter(`error`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, keep := f([]byte("fatal error occurred"))
	if !keep {
		t.Error("expected chunk to be kept")
	}
	if string(out) != "fatal error occurred" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestGrepFilterNoMatch(t *testing.T) {
	f, err := filter.GrepFilter(`error`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, keep := f([]byte("everything is fine"))
	if keep {
		t.Error("expected chunk to be dropped")
	}
}

func TestGrepFilterInvalidPattern(t *testing.T) {
	_, err := filter.GrepFilter(`[invalid`)
	if err == nil {
		t.Error("expected error for invalid regex")
	}
}

func TestTrimFilterRemovesWhitespace(t *testing.T) {
	f := filter.TrimFilter()
	out, keep := f([]byte("  hello world  \n"))
	if !keep {
		t.Error("expected chunk to be kept")
	}
	if string(out) != "hello world" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestTrimFilterDropsEmpty(t *testing.T) {
	f := filter.TrimFilter()
	_, keep := f([]byte("   \n  "))
	if keep {
		t.Error("expected empty chunk to be dropped")
	}
}

func TestMaxSizeFilter(t *testing.T) {
	f := filter.MaxSizeFilter(10)
	_, keep := f([]byte("this is too long for the limit"))
	if keep {
		t.Error("expected oversized chunk to be dropped")
	}
	out, keep := f([]byte("short"))
	if !keep {
		t.Error("expected small chunk to be kept")
	}
	if string(out) != "short" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestChain(t *testing.T) {
	grep, _ := filter.GrepFilter(`data`)
	trim := filter.TrimFilter()
	chained := filter.Chain(trim, grep)

	out, keep := chained([]byte("  data packet  "))
	if !keep {
		t.Error("expected chained filter to keep chunk")
	}
	if string(out) != "data packet" {
		t.Errorf("unexpected output: %q", out)
	}

	_, keep = chained([]byte("  noise  "))
	if keep {
		t.Error("expected chained filter to drop chunk")
	}
}
