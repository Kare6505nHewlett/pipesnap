// Package offset provides utilities for tracking and seeking to byte/chunk
// offsets within a snapshot file, enabling random-access replay.
package offset

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/user/pipesnap/internal/header"
)

// Entry records the byte position and chunk index of a single chunk.
type Entry struct {
	ChunkIndex int64
	ByteOffset int64
}

// Build scans a snapshot file and returns an ordered slice of Entries,
// one per chunk, recording the byte offset at which each chunk begins.
func Build(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("offset.Build: open %s: %w", path, err)
	}
	defer f.Close()

	// Skip the file-level header written by internal/header.
	if _, err := header.Read(f); err != nil {
		return nil, fmt.Errorf("offset.Build: read header: %w", err)
	}

	var entries []Entry
	var idx int64

	for {
		pos, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("offset.Build: seek: %w", err)
		}

		var size uint32
		if err := binary.Read(f, binary.LittleEndian, &size); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("offset.Build: read chunk size: %w", err)
		}

		entries = append(entries, Entry{ChunkIndex: idx, ByteOffset: pos})
		idx++

		if _, err := f.Seek(int64(size), io.SeekCurrent); err != nil {
			return nil, fmt.Errorf("offset.Build: skip chunk body: %w", err)
		}
	}

	return entries, nil
}

// SeekTo positions r to the start of the chunk at the given byte offset.
// r must be an io.ReadSeeker (e.g. *os.File).
func SeekTo(r io.ReadSeeker, byteOffset int64) error {
	if _, err := r.Seek(byteOffset, io.SeekStart); err != nil {
		return fmt.Errorf("offset.SeekTo: %w", err)
	}
	return nil
}
