package snapshot

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Meta holds metadata written at the start of a snapshot file.
type Meta struct {
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Label     string    `json:"label,omitempty"`
}

// Writer captures stdin chunks and writes them to a gzip-compressed snapshot file.
type Writer struct {
	file *os.File
	gz   *gzip.Writer
}

// NewWriter creates a snapshot file at path and writes the header metadata.
func NewWriter(path, label string) (*Writer, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("snapshot: create file: %w", err)
	}
	gz := gzip.NewWriter(f)
	meta := Meta{Version: 1, CreatedAt: time.Now().UTC(), Label: label}
	enc := json.NewEncoder(gz)
	if err := enc.Encode(meta); err != nil {
		f.Close()
		return nil, fmt.Errorf("snapshot: write meta: %w", err)
	}
	return &Writer{file: f, gz: gz}, nil
}

// Write copies p into the snapshot stream.
func (w *Writer) Write(p []byte) (int, error) {
	return w.gz.Write(p)
}

// Close flushes and closes the underlying gzip and file handles.
func (w *Writer) Close() error {
	if err := w.gz.Close(); err != nil {
		w.file.Close()
		return fmt.Errorf("snapshot: close gzip: %w", err)
	}
	return w.file.Close()
}

// OpenReader opens a snapshot file and returns the Meta and a reader for the raw stream.
func OpenReader(path string) (Meta, io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return Meta{}, nil, fmt.Errorf("snapshot: open file: %w", err)
	}
	gz, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return Meta{}, nil, fmt.Errorf("snapshot: open gzip: %w", err)
	}
	var meta Meta
	dec := json.NewDecoder(gz)
	if err := dec.Decode(&meta); err != nil {
		gz.Close()
		f.Close()
		return Meta{}, nil, fmt.Errorf("snapshot: read meta: %w", err)
	}
	// Return a combined closer so callers only need one Close call.
	return meta, &multiCloser{Reader: io.MultiReader(dec.Buffered(), gz), closers: []io.Closer{gz, f}}, nil
}

type multiCloser struct {
	io.Reader
	closers []io.Closer
}

func (m *multiCloser) Close() error {
	var first error
	for _, c := range m.closers {
		if err := c.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
