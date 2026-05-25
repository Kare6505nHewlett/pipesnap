// Package buffer provides a fixed-capacity in-memory chunk buffer that
// accumulates writes and flushes to a downstream writer once the buffer
// reaches its configured size or Flush is called explicitly.
package buffer

import (
	"errors"
	"io"
)

// ErrCapacity is returned when cap is less than 1.
var ErrCapacity = errors.New("buffer: capacity must be at least 1")

// Buffer accumulates chunks and flushes them in a single batch.
type Buffer struct {
	cap    int
	chunks [][]byte
	bytes  int
	dst    io.Writer
}

// New creates a Buffer that flushes to dst once cap chunks have been
// accumulated. cap must be >= 1.
func New(cap int, dst io.Writer) (*Buffer, error) {
	if cap < 1 {
		return nil, ErrCapacity
	}
	if dst == nil {
		return nil, errors.New("buffer: dst must not be nil")
	}
	return &Buffer{
		cap:    cap,
		chunks: make([][]byte, 0, cap),
		dst:    dst,
	}, nil
}

// Write appends p to the buffer. When the buffer reaches capacity it is
// flushed automatically. The full length of p and a nil error are returned
// unless the flush fails.
func (b *Buffer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	chunk := make([]byte, len(p))
	copy(chunk, p)
	b.chunks = append(b.chunks, chunk)
	b.bytes += len(chunk)
	if len(b.chunks) >= b.cap {
		if err := b.Flush(); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// Flush writes all buffered chunks to dst and resets the buffer.
func (b *Buffer) Flush() error {
	for _, chunk := range b.chunks {
		if _, err := b.dst.Write(chunk); err != nil {
			return err
		}
	}
	b.chunks = b.chunks[:0]
	b.bytes = 0
	return nil
}

// Len returns the number of chunks currently held in the buffer.
func (b *Buffer) Len() int { return len(b.chunks) }

// Bytes returns the total byte size of all chunks currently in the buffer.
func (b *Buffer) Bytes() int { return b.bytes }
