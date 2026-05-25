// Package fanout distributes each incoming chunk to multiple named snapshot
// destinations, applying an optional per-destination filter before writing.
package fanout

import (
	"fmt"
	"io"
)

// Destination is a single named output for the fanout writer.
type Destination struct {
	// Name is used in error messages.
	Name string
	// Dst is the underlying writer that receives chunks.
	Dst io.Writer
	// Filter, when non-nil, is called with each chunk. Returning false drops
	// the chunk for this destination only.
	Filter func([]byte) bool
}

// Fanout writes each chunk to all registered destinations.
type Fanout struct {
	dests []Destination
}

// New creates a Fanout with the supplied destinations. At least one destination
// is required.
func New(dests []Destination) (*Fanout, error) {
	if len(dests) == 0 {
		return nil, fmt.Errorf("fanout: at least one destination required")
	}
	for i, d := range dests {
		if d.Dst == nil {
			return nil, fmt.Errorf("fanout: destination %d (%q) has nil writer", i, d.Name)
		}
	}
	return &Fanout{dests: dests}, nil
}

// Write sends p to every destination whose filter (if any) accepts it.
// All destinations are attempted; the first error encountered is returned
// after all writes complete.
func (f *Fanout) Write(p []byte) (int, error) {
	var firstErr error
	for _, d := range f.dests {
		if d.Filter != nil && !d.Filter(p) {
			continue
		}
		if _, err := d.Dst.Write(p); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("fanout: destination %q: %w", d.Name, err)
		}
	}
	return len(p), firstErr
}

// Len returns the number of registered destinations.
func (f *Fanout) Len() int { return len(f.dests) }
