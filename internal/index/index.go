// Package index builds and queries a lightweight offset index over snapshot
// files, allowing callers to seek directly to a chunk by sequence number
// without scanning from the beginning of the file.
package index

import (
	"encoding/json"
	"fmt"
	"os"
)

// Entry records the byte offset and size of a single chunk within a snapshot.
type Entry struct {
	Seq    int   `json:"seq"`
	Offset int64 `json:"offset"`
	Size   int   `json:"size"`
}

// Index is an ordered list of chunk entries for a snapshot file.
type Index struct {
	Entries []Entry `json:"entries"`
}

// Add appends a new entry to the index.
func (idx *Index) Add(seq int, offset int64, size int) {
	idx.Entries = append(idx.Entries, Entry{Seq: seq, Offset: offset, Size: size})
}

// Lookup returns the Entry for the given sequence number, or an error if not
// found.
func (idx *Index) Lookup(seq int) (Entry, error) {
	for _, e := range idx.Entries {
		if e.Seq == seq {
			return e, nil
		}
	}
	return Entry{}, fmt.Errorf("index: seq %d not found", seq)
}

// Len returns the number of indexed chunks.
func (idx *Index) Len() int {
	return len(idx.Entries)
}

// Save writes the index to path as JSON.
func Save(path string, idx *Index) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("index: create %s: %w", path, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(idx)
}

// Load reads an index from path.
func Load(path string) (*Index, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("index: open %s: %w", path, err)
	}
	defer f.Close()
	var idx Index
	if err := json.NewDecoder(f).Decode(&idx); err != nil {
		return nil, fmt.Errorf("index: decode %s: %w", path, err)
	}
	return &idx, nil
}

// PathFor returns the conventional index path for a snapshot file.
func PathFor(snapPath string) string {
	return snapPath + ".idx"
}
