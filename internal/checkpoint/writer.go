package checkpoint

import (
	"io"
	"time"
)

// Writer wraps an io.Writer and updates a checkpoint state after each write.
type Writer struct {
	inner        io.Writer
	checkpointPath string
	state        State
}

// NewWriter creates a Writer that persists checkpoint state to checkpointPath
// after every chunk written to inner.
func NewWriter(inner io.Writer, checkpointPath string, initial State) *Writer {
	if initial.CreatedAt.IsZero() {
		initial.CreatedAt = time.Now()
	}
	return &Writer{
		inner:          inner,
		checkpointPath: checkpointPath,
		state:          initial,
	}
}

// Write passes p to the underlying writer and updates the checkpoint on success.
func (w *Writer) Write(p []byte) (int, error) {
	n, err := w.inner.Write(p)
	if err != nil {
		return n, err
	}
	w.state.BytesWritten += int64(n)
	w.state.ChunksWritten++
	if saveErr := Save(w.checkpointPath, w.state); saveErr != nil {
		return n, saveErr
	}
	return n, nil
}

// State returns the current checkpoint state.
func (w *Writer) State() State {
	return w.state
}
