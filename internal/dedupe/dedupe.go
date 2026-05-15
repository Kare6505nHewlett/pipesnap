// Package dedupe provides chunk deduplication for snapshot streams.
// It tracks seen chunks by hash and drops repeated content.
package dedupe

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// Filter drops chunks whose content has already been seen.
type Filter struct {
	seen map[string]struct{}
}

// New returns a new deduplication Filter with an empty seen set.
func New() *Filter {
	return &Filter{seen: make(map[string]struct{})}
}

// IsDuplicate returns true if the chunk has been seen before.
// If not seen, it records the chunk and returns false.
func (f *Filter) IsDuplicate(chunk []byte) bool {
	h := hash(chunk)
	if _, ok := f.seen[h]; ok {
		return true
	}
	f.seen[h] = struct{}{}
	return false
}

// Reset clears all previously seen hashes.
func (f *Filter) Reset() {
	f.seen = make(map[string]struct{})
}

// Len returns the number of unique chunks seen so far.
func (f *Filter) Len() int {
	return len(f.seen)
}

// NewWriter wraps dst and writes only chunks not seen before.
func NewWriter(dst io.Writer, f *Filter) io.Writer {
	return &dedupeWriter{dst: dst, filter: f}
}

type dedupeWriter struct {
	dst    io.Writer
	filter *Filter
}

func (w *dedupeWriter) Write(p []byte) (int, error) {
	if w.filter.IsDuplicate(p) {
		return len(p), nil
	}
	return w.dst.Write(p)
}

func hash(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
