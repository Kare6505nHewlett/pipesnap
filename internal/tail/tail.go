// Package tail provides utilities for reading the last N chunks
// from a snapshot file, useful for inspecting recent pipeline activity.
package tail

import (
	"fmt"
	"io"

	"github.com/user/pipesnap/internal/snapshot"
)

// Options configures the tail operation.
type Options struct {
	// N is the number of chunks to return from the end of the snapshot.
	// If N <= 0, all chunks are returned.
	N int
}

// Chunk holds the raw bytes of a single snapshot chunk.
type Chunk struct {
	Index int
	Data  []byte
}

// Read opens the snapshot at path and returns up to opts.N chunks from
// the end of the stream. If opts.N <= 0 all chunks are returned.
func Read(path string, opts Options) ([]Chunk, error) {
	r, err := snapshot.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("tail: open snapshot: %w", err)
	}
	defer r.Close()

	var all []Chunk
	index := 0
	for {
		buf, err := r.ReadChunk()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tail: read chunk: %w", err)
		}
		all = append(all, Chunk{Index: index, Data: buf})
		index++
	}

	if opts.N <= 0 || opts.N >= len(all) {
		return all, nil
	}
	return all[len(all)-opts.N:], nil
}

// Write copies chunks to w, respecting the original order.
func Write(w io.Writer, chunks []Chunk) error {
	for _, c := range chunks {
		if _, err := w.Write(c.Data); err != nil {
			return fmt.Errorf("tail: write chunk %d: %w", c.Index, err)
		}
	}
	return nil
}
