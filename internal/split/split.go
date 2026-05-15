// Package split provides utilities for splitting a snapshot file into
// multiple smaller snapshot files based on chunk count or byte size.
package split

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourorg/pipesnap/internal/snapshot"
)

// Options controls how a snapshot is split.
type Options struct {
	// MaxChunks is the maximum number of chunks per output file.
	// Zero means no chunk-based splitting.
	MaxChunks int
	// MaxBytes is the maximum number of payload bytes per output file.
	// Zero means no byte-based splitting.
	MaxBytes int64
	// OutDir is the directory where split files are written.
	OutDir string
	// Prefix is prepended to each output filename.
	Prefix string
}

// File splits the snapshot at srcPath into multiple files according to opts.
// It returns the paths of the files that were created.
func File(srcPath string, opts Options) ([]string, error) {
	if opts.MaxChunks <= 0 && opts.MaxBytes <= 0 {
		return nil, fmt.Errorf("split: at least one of MaxChunks or MaxBytes must be set")
	}
	if opts.OutDir == "" {
		opts.OutDir = filepath.Dir(srcPath)
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return nil, fmt.Errorf("split: mkdir: %w", err)
	}

	r, err := snapshot.OpenReader(srcPath)
	if err != nil {
		return nil, fmt.Errorf("split: open source: %w", err)
	}
	defer r.Close()

	var (
		paths      []string
		partIndex  int
		chunkCount int
		byteCount  int64
		w          *snapshot.Writer
	)

	newPart := func() error {
		if w != nil {
			if err := w.Close(); err != nil {
				return err
			}
		}
		name := fmt.Sprintf("%s%04d.snap", opts.Prefix, partIndex)
		p := filepath.Join(opts.OutDir, name)
		paths = append(paths, p)
		w, err = snapshot.NewWriter(p)
		if err != nil {
			return fmt.Errorf("split: create part %d: %w", partIndex, err)
		}
		partIndex++
		chunkCount = 0
		byteCount = 0
		return nil
	}

	if err := newPart(); err != nil {
		return nil, err
	}

	for {
		chunk, err := r.Next()
		if err != nil {
			break
		}
		needRotate := (opts.MaxChunks > 0 && chunkCount >= opts.MaxChunks) ||
			(opts.MaxBytes > 0 && byteCount+int64(len(chunk)) > opts.MaxBytes && chunkCount > 0)
		if needRotate {
			if err := newPart(); err != nil {
				return nil, err
			}
		}
		if _, werr := w.Write(chunk); werr != nil {
			return nil, fmt.Errorf("split: write chunk: %w", werr)
		}
		chunkCount++
		byteCount += int64(len(chunk))
	}

	if w != nil {
		if err := w.Close(); err != nil {
			return nil, fmt.Errorf("split: close final part: %w", err)
		}
	}
	return paths, nil
}
