package rotate

import (
	"fmt"
	"io"
	"os"
)

// Writer wraps an underlying snapshot writer and triggers rotation after
// each completed snapshot file is closed.
type Writer struct {
	manager *Manager
	current io.WriteCloser
	path    string
	bytesWritten int64
}

// NewWriter opens a new snapshot file via the Manager and returns a Writer.
// Rotation is applied before opening the new file.
func NewWriter(m *Manager) (*Writer, error) {
	if _, err := m.Rotate(); err != nil {
		return nil, fmt.Errorf("rotate writer: pre-rotate: %w", err)
	}

	path := m.NextPath()
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("rotate writer: create %s: %w", path, err)
	}

	return &Writer{
		manager: m,
		current: f,
		path:    path,
	}, nil
}

// Write writes p to the current snapshot file.
func (w *Writer) Write(p []byte) (int, error) {
	n, err := w.current.Write(p)
	w.bytesWritten += int64(n)
	return n, err
}

// Close closes the underlying file.
func (w *Writer) Close() error {
	return w.current.Close()
}

// Path returns the path of the currently open snapshot file.
func (w *Writer) Path() string {
	return w.path
}

// BytesWritten returns the number of bytes written to the current file.
func (w *Writer) BytesWritten() int64 {
	return w.bytesWritten
}
