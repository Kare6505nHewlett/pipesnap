package stats

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestPrinterContainsFields(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	s := Summary{
		Elapsed: 2 * time.Second,
		Bytes:   1024,
		Chunks:  8,
		Dropped: 2,
	}
	p.Print(s)
	out := buf.String()
	for _, want := range []string{"chunks", "bytes", "dropped", "elapsed", "drop %"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestPrinterDropPercentage(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	s := Summary{
		Elapsed: time.Second,
		Bytes:   500,
		Chunks:  1,
		Dropped: 1,
	}
	p.Print(s)
	out := buf.String()
	if !strings.Contains(out, "50.0%") {
		t.Errorf("expected 50.0%% in output, got:\n%s", out)
	}
}

func TestPrinterNoDropPercentageWhenEmpty(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	s := Summary{}
	p.Print(s)
	out := buf.String()
	if strings.Contains(out, "drop %") {
		t.Errorf("expected no drop %% line when no chunks, got:\n%s", out)
	}
}
