// Package merge provides utilities for combining multiple snapshot files
// into a single ordered stream, useful for replaying distributed pipeline captures.
package merge

import (
	"fmt"
	"io"
	"sort"

	"github.com/user/pipesnap/internal/snapshot"
)

// Entry holds a decoded chunk along with its source path.
type Entry struct {
	Source string
	Data   []byte
}

// Merge reads all chunks from each snapshot path in order and writes them
// sequentially to dst. Snapshots are processed in the order provided.
// Returns the total number of bytes written across all chunks.
func Merge(dst io.Writer, paths []string) (int64, error) {
	if len(paths) == 0 {
		return 0, fmt.Errorf("merge: no snapshot paths provided")
	}

	var total int64
	for _, p := range paths {
		n, err := drainSnapshot(dst, p)
		total += n
		if err != nil {
			return total, fmt.Errorf("merge: reading %q: %w", p, err)
		}
	}
	return total, nil
}

// Collect reads all chunks from each snapshot path and returns them as a
// slice of Entry values. Entries are sorted by source path to ensure
// deterministic ordering when paths share a common prefix.
func Collect(paths []string) ([]Entry, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("merge: no snapshot paths provided")
	}

	sorted := make([]string, len(paths))
	copy(sorted, paths)
	sort.Strings(sorted)

	var entries []Entry
	for _, p := range sorted {
		r, err := snapshot.OpenReader(p)
		if err != nil {
			return nil, fmt.Errorf("merge: open %q: %w", p, err)
		}
		for {
			chunk, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("merge: read %q: %w", p, err)
			}
			entries = append(entries, Entry{Source: p, Data: chunk})
		}
	}
	return entries, nil
}

func drainSnapshot(dst io.Writer, path string) (int64, error) {
	r, err := snapshot.OpenReader(path)
	if err != nil {
		return 0, err
	}
	var total int64
	for {
		chunk, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return total, err
		}
		n, werr := dst.Write(chunk)
		total += int64(n)
		if werr != nil {
			return total, werr
		}
	}
	return total, nil
}
