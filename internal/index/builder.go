package index

import (
	"fmt"
	"io"

	"github.com/user/pipesnap/internal/header"
)

// Builder scans a snapshot file and constructs an Index from its chunk headers.
type Builder struct {
	r   io.ReadSeeker
	idx *Index
}

// NewBuilder creates a Builder that reads from r.
func NewBuilder(r io.ReadSeeker) *Builder {
	return &Builder{r: r, idx: &Index{}}
}

// Build scans r from the current position, reading chunk headers and recording
// each chunk's offset and payload size. Returns the completed Index.
func (b *Builder) Build() (*Index, error) {
	seq := 0
	for {
		offset, err := b.r.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("index builder: seek: %w", err)
		}

		hdr, err := header.Read(b.r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("index builder: read header at seq %d: %w", seq, err)
		}

		b.idx.Add(seq, offset, hdr.Size)

		// Skip over the payload bytes.
		if _, err := io.CopyN(io.Discard, b.r, int64(hdr.Size)); err != nil {
			return nil, fmt.Errorf("index builder: skip payload seq %d: %w", seq, err)
		}

		seq++
	}
	return b.idx, nil
}

// BuildFile opens snapPath, builds an index, saves it to the conventional
// sidecar path, and returns the index.
func BuildFile(snapPath string) (*Index, error) {
	import_os := openFile // resolved below via closure to keep import clean
	f, err := import_os(snapPath)
	if err != nil {
		return nil, fmt.Errorf("index: open snapshot: %w", err)
	}
	defer f.Close()

	idx, err := NewBuilder(f).Build()
	if err != nil {
		return nil, err
	}

	idxPath := PathFor(snapPath)
	if err := Save(idxPath, idx); err != nil {
		return nil, err
	}
	return idx, nil
}
