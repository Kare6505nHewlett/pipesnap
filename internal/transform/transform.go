// Package transform provides line-level transformation functions
// that can be applied to snapshot chunks during replay or capture.
package transform

import (
	"bytes"
	"strings"
	"unicode"
)

// Func is a transformation applied to a single chunk of bytes.
// It returns the transformed bytes and whether the chunk should be kept.
type Func func([]byte) ([]byte, bool)

// Chain applies a sequence of Funcs in order, short-circuiting if any drops the chunk.
func Chain(fns ...Func) Func {
	return func(b []byte) ([]byte, bool) {
		for _, fn := range fns {
			var keep bool
			b, keep = fn(b)
			if !keep {
				return nil, false
			}
		}
		return b, true
	}
}

// UpperCase transforms every byte in the chunk to upper-case.
func UpperCase(b []byte) ([]byte, bool) {
	return bytes.ToUpper(b), true
}

// LowerCase transforms every byte in the chunk to lower-case.
func LowerCase(b []byte) ([]byte, bool) {
	return bytes.ToLower(b), true
}

// ReplaceAll replaces all occurrences of old with new inside the chunk.
func ReplaceAll(old, new string) Func {
	return func(b []byte) ([]byte, bool) {
		replaced := strings.ReplaceAll(string(b), old, new)
		return []byte(replaced), true
	}
}

// StripControl removes non-printable, non-space control characters from the chunk.
func StripControl(b []byte) ([]byte, bool) {
	filtered := bytes.Map(func(r rune) rune {
		if unicode.IsControl(r) && !unicode.IsSpace(r) {
			return -1
		}
		return r
	}, b)
	return filtered, len(filtered) > 0
}

// Truncate limits the chunk to at most n bytes, preserving a trailing newline if present.
func Truncate(n int) Func {
	return func(b []byte) ([]byte, bool) {
		if n <= 0 || len(b) <= n {
			return b, true
		}
		trunc := b[:n]
		if b[len(b)-1] == '\n' {
			trunc = append(trunc, '\n')
		}
		return trunc, true
	}
}
