// Package patch applies a set of byte-level substitutions to snapshot chunks.
// Each substitution replaces all occurrences of a literal byte sequence with
// a replacement sequence, leaving unmatched bytes untouched.
package patch

import (
	"bytes"
	"fmt"
)

// Sub describes a single find-and-replace substitution.
type Sub struct {
	Find    []byte
	Replace []byte
}

// Patcher holds an ordered list of substitutions.
type Patcher struct {
	subs []Sub
}

// New creates a Patcher from the provided substitutions.
// An error is returned if any Sub has a nil or empty Find slice.
func New(subs []Sub) (*Patcher, error) {
	for i, s := range subs {
		if len(s.Find) == 0 {
			return nil, &ErrEmptyFind{Index: i}
		}
	}
	return &Patcher{subs: subs}, nil
}

// Apply runs all substitutions against data in order and returns the result.
func (p *Patcher) Apply(data []byte) []byte {
	for _, s := range p.subs {
		data = bytes.ReplaceAll(data, s.Find, s.Replace)
	}
	return data
}

// Len returns the number of substitutions registered in the Patcher.
func (p *Patcher) Len() int { return len(p.subs) }

// ErrEmptyFind is returned when a Sub has an empty Find field.
type ErrEmptyFind struct {
	Index int
}

func (e *ErrEmptyFind) Error() string {
	return fmt.Sprintf("patch: sub[%d]: Find must not be empty", e.Index)
}
