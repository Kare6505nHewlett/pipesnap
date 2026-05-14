// Package truncate provides utilities for truncating snapshot files
// to a maximum number of chunks or bytes, discarding the oldest entries.
package truncate

import (
	"fmt"
	"os"

	"github.com/user/pipesnap/internal/snapshot"
)

// Options controls how truncation is applied.
type Options struct {
	// MaxChunks is the maximum number of chunks to retain (0 = unlimited).
	MaxChunks int
	// MaxBytes is the maximum total payload bytes to retain (0 = unlimited).
	MaxBytes int64
}

// File reads the snapshot at src, discards leading chunks until the file
// satisfies opts, and writes the result to dst. src and dst may be the same
// path; in that case a temp file is used as an intermediary.
func File(src, dst string, opts Options) error {
	if opts.MaxChunks == 0 && opts.MaxBytes == 0 {
		return nil
	}

	chunks, err := readAll(src)
	if err != nil {
		return fmt.Errorf("truncate: read %s: %w", src, err)
	}

	chunks = apply(chunks, opts)

	tmp := dst + ".tmp"
	if err := writeAll(tmp, chunks); err != nil {
		return fmt.Errorf("truncate: write %s: %w", tmp, err)
	}

	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("truncate: rename: %w", err)
	}
	return nil
}

// apply drops leading chunks until the slice satisfies opts.
func apply(chunks [][]byte, opts Options) [][]byte {
	if opts.MaxChunks > 0 && len(chunks) > opts.MaxChunks {
		chunks = chunks[len(chunks)-opts.MaxChunks:]
	}
	if opts.MaxBytes > 0 {
		var total int64
		for _, c := range chunks {
			total += int64(len(c))
		}
		for total > opts.MaxBytes && len(chunks) > 0 {
			total -= int64(len(chunks[0]))
			chunks = chunks[1:]
		}
	}
	return chunks
}

func readAll(path string) ([][]byte, error) {
	r, err := snapshot.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var out [][]byte
	for {
		b, err := r.ReadChunk()
		if err != nil {
			break
		}
		out = append(out, b)
	}
	return out, nil
}

func writeAll(path string, chunks [][]byte) error {
	w, err := snapshot.NewWriter(path)
	if err != nil {
		return err
	}
	for _, c := range chunks {
		if _, err := w.Write(c); err != nil {
			_ = w.Close()
			return err
		}
	}
	return w.Close()
}
