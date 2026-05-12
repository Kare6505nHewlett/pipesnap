// Package filter provides utilities to filter and transform snapshot chunks
// before writing or after reading, enabling lightweight preprocessing pipelines.
package filter

import (
	"bytes"
	"regexp"
)

// Func is a function that transforms a chunk of bytes.
// It returns the transformed bytes and whether the chunk should be kept.
type Func func(chunk []byte) ([]byte, bool)

// Chain combines multiple filter functions into one, applying them in order.
// If any filter drops the chunk (returns false), the chain stops.
func Chain(filters ...Func) Func {
	return func(chunk []byte) ([]byte, bool) {
		current := chunk
		for _, f := range filters {
			var keep bool
			current, keep = f(current)
			if !keep {
				return nil, false
			}
		}
		return current, true
	}
}

// GrepFilter returns a Func that keeps only chunks matching the given pattern.
func GrepFilter(pattern string) (Func, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return func(chunk []byte) ([]byte, bool) {
		return chunk, re.Match(chunk)
	}, nil
}

// TrimFilter returns a Func that trims leading and trailing whitespace from each chunk.
func TrimFilter() Func {
	return func(chunk []byte) ([]byte, bool) {
		trimmed := bytes.TrimSpace(chunk)
		if len(trimmed) == 0 {
			return nil, false
		}
		return trimmed, true
	}
}

// MaxSizeFilter returns a Func that drops chunks exceeding maxBytes in size.
func MaxSizeFilter(maxBytes int) Func {
	return func(chunk []byte) ([]byte, bool) {
		if len(chunk) > maxBytes {
			return nil, false
		}
		return chunk, true
	}
}
