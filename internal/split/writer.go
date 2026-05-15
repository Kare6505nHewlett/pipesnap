package split

import (
	"fmt"

	"github.com/yourorg/pipesnap/internal/snapshot"
)

// Writer is an io.Writer that transparently splits incoming chunks across
// multiple snapshot files according to the supplied Options.
type Writer struct {
	opts      Options
	current   *snapshot.Writer
	partIndex int
	chunks    int
	bytes     int64
	Paths     []string
}

// NewWriter creates a split.Writer. The first part file is opened immediately.
func NewWriter(opts Options) (*Writer, error) {
	if opts.MaxChunks <= 0 && opts.MaxBytes <= 0 {
		return nil, fmt.Errorf("split.NewWriter: at least one of MaxChunks or MaxBytes must be set")
	}
	sw := &Writer{opts: opts}
	if err := sw.openNext(); err != nil {
		return nil, err
	}
	return sw, nil
}

func (sw *Writer) openNext() error {
	if sw.current != nil {
		if err := sw.current.Close(); err != nil {
			return err
		}
	}
	name := fmt.Sprintf("%s%04d.snap", sw.opts.Prefix, sw.partIndex)
	import_path := sw.opts.OutDir + "/" + name
	w, err := snapshot.NewWriter(import_path)
	if err != nil {
		return fmt.Errorf("split.Writer: open part %d: %w", sw.partIndex, err)
	}
	sw.Paths = append(sw.Paths, import_path)
	sw.current = w
	sw.partIndex++
	sw.chunks = 0
	sw.bytes = 0
	return nil
}

// Write writes p as a single chunk, rotating to a new part file if needed.
func (sw *Writer) Write(p []byte) (int, error) {
	needRotate := (sw.opts.MaxChunks > 0 && sw.chunks >= sw.opts.MaxChunks) ||
		(sw.opts.MaxBytes > 0 && sw.bytes+int64(len(p)) > sw.opts.MaxBytes && sw.chunks > 0)
	if needRotate {
		if err := sw.openNext(); err != nil {
			return 0, err
		}
	}
	n, err := sw.current.Write(p)
	if err != nil {
		return n, err
	}
	sw.chunks++
	sw.bytes += int64(n)
	return n, nil
}

// Close flushes and closes the current part file.
func (sw *Writer) Close() error {
	if sw.current == nil {
		return nil
	}
	err := sw.current.Close()
	sw.current = nil
	return err
}
