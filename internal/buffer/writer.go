package buffer

import (
	"errors"
	"io"
)

// Writer wraps a Buffer and implements io.WriteCloser. Closing the writer
// flushes any remaining buffered chunks to the underlying destination.
type Writer struct {
	buf *Buffer
	dst io.WriteCloser
}

// NewWriter creates a Writer that buffers up to cap chunks before flushing
// to dst. dst is closed when the Writer is closed.
func NewWriter(cap int, dst io.WriteCloser) (*Writer, error) {
	if dst == nil {
		return nil, errors.New("buffer: dst must not be nil")
	}
	buf, err := New(cap, dst)
	if err != nil {
		return nil, err
	}
	return &Writer{buf: buf, dst: dst}, nil
}

// Write buffers p. An automatic flush occurs when the buffer reaches capacity.
func (w *Writer) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

// Close flushes any remaining chunks and closes the underlying writer.
func (w *Writer) Close() error {
	if err := w.buf.Flush(); err != nil {
		return err
	}
	return w.dst.Close()
}
