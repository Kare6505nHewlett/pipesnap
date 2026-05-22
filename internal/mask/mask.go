// Package mask provides chunk-level data masking for sensitive fields
// in snapshot streams. It supports redacting substrings that match
// registered patterns, replacing them with a fixed placeholder.
package mask

import (
	"fmt"
	"io"
	"regexp"
)

// Masker replaces sensitive patterns in chunk data with a placeholder.
type Masker struct {
	patterns     []*regexp.Regexp
	placeholder  []byte
}

// New creates a Masker that replaces any match of the given regex patterns
// with placeholder. Returns an error if any pattern fails to compile.
func New(placeholder string, patterns ...string) (*Masker, error) {
	if len(patterns) == 0 {
		return nil, fmt.Errorf("mask: at least one pattern is required")
	}
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("mask: invalid pattern %q: %w", p, err)
		}
		compiled = append(compiled, re)
	}
	return &Masker{
		patterns:    compiled,
		placeholder: []byte(placeholder),
	}, nil
}

// Apply returns a copy of data with all pattern matches replaced by the
// placeholder. The original slice is never modified.
func (m *Masker) Apply(data []byte) []byte {
	out := data
	for _, re := range m.patterns {
		out = re.ReplaceAll(out, m.placeholder)
	}
	return out
}

// NewWriter wraps dst so that every Write call passes the chunk through the
// Masker before forwarding it to dst.
func NewWriter(dst io.Writer, m *Masker) io.Writer {
	return &maskWriter{dst: dst, masker: m}
}

type maskWriter struct {
	dst    io.Writer
	masker *Masker
}

func (w *maskWriter) Write(p []byte) (int, error) {
	masked := w.masker.Apply(p)
	_, err := w.dst.Write(masked)
	// Return the original length so callers do not see a short-write error
	// even when the masked output is shorter than the input.
	return len(p), err
}
