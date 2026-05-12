package filter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pipesnap/pipesnap/internal/filter"
)

func TestWriterPassthrough(t *testing.T) {
	var buf bytes.Buffer
	passthrough := func(chunk []byte) ([]byte, bool) { return chunk, true }
	w := filter.NewWriter(&buf, passthrough)

	_, err := w.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "hello" {
		t.Errorf("expected 'hello', got %q", buf.String())
	}
}

func TestWriterDropsChunk(t *testing.T) {
	var buf bytes.Buffer
	drop := func(chunk []byte) ([]byte, bool) { return nil, false }
	w := filter.NewWriter(&buf, drop)

	n, err := w.Write([]byte("should be dropped"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len("should be dropped") {
		t.Errorf("expected n=%d, got %d", len("should be dropped"), n)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty buffer, got %q", buf.String())
	}
}

func TestWriterWithGrepFilter(t *testing.T) {
	var buf bytes.Buffer
	grep, _ := filter.GrepFilter(`WARN|ERROR`)
	w := filter.NewWriter(&buf, grep)

	lines := []string{
		"INFO starting service\n",
		"WARN disk usage high\n",
		"DEBUG checkpoint reached\n",
		"ERROR connection lost\n",
	}
	for _, line := range lines {
		w.Write([]byte(line)) //nolint:errcheck
	}

	result := buf.String()
	if !strings.Contains(result, "WARN disk usage high") {
		t.Errorf("expected WARN line in output, got: %q", result)
	}
	if !strings.Contains(result, "ERROR connection lost") {
		t.Errorf("expected ERROR line in output, got: %q", result)
	}
	if strings.Contains(result, "INFO") || strings.Contains(result, "DEBUG") {
		t.Errorf("unexpected INFO/DEBUG lines in output: %q", result)
	}
}
