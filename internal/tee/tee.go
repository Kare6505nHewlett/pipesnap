// Package tee provides an io.Writer that writes to multiple destinations
// simultaneously, similar to the Unix tee command. It is used in pipesnap
// to fan out snapshot data to both a file and downstream consumers.
package tee

import (
	"fmt"
	"io"
)

// Writer writes each chunk to all registered destinations.
// If any destination returns an error the write is aborted and the
// error is returned, but destinations that were written before the
// failure are not rolled back.
type Writer struct {
	dests []io.Writer
}

// New creates a Writer that fans out to each of the supplied destinations.
// At least one destination must be provided.
func New(dests ...io.Writer) (*Writer, error) {
	if len(dests) == 0 {
		return nil, fmt.Errorf("tee: at least one destination is required")
	}
	return &Writer{dests: dests}, nil
}

// Write writes p to every destination in order. The first error encountered
// is returned immediately; subsequent destinations are not written.
func (w *Writer) Write(p []byte) (int, error) {
	for _, d := range w.dests {
		n, err := d.Write(p)
		if err != nil {
			return n, fmt.Errorf("tee: write error: %w", err)
		}
		if n != len(p) {
			return n, io.ErrShortWrite
		}
	}
	return len(p), nil
}

// Add appends a new destination to the writer. Writes after this call will
// also be sent to d.
func (w *Writer) Add(d io.Writer) {
	w.dests = append(w.dests, d)
}

// Len returns the number of destinations currently registered.
func (w *Writer) Len() int {
	return len(w.dests)
}
