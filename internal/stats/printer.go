package stats

import (
	"fmt"
	"io"
)

// Printer writes human-readable statistics to a writer.
type Printer struct {
	out io.Writer
}

// NewPrinter creates a Printer that writes to out.
func NewPrinter(out io.Writer) *Printer {
	return &Printer{out: out}
}

// Print formats and writes the given Summary to the printer's output.
func (p *Printer) Print(s Summary) {
	fmt.Fprintf(p.out, "--- pipesnap stats ---\n")
	fmt.Fprintf(p.out, "elapsed : %s\n", s.Elapsed.Round(1*1000000))
	fmt.Fprintf(p.out, "chunks  : %d\n", s.Chunks)
	fmt.Fprintf(p.out, "bytes   : %d\n", s.Bytes)
	fmt.Fprintf(p.out, "dropped : %d\n", s.Dropped)
	if s.Chunks+s.Dropped > 0 {
		total := s.Chunks + s.Dropped
		pct := float64(s.Dropped) / float64(total) * 100.0
		fmt.Fprintf(p.out, "drop %%  : %.1f%%\n", pct)
	}
}
