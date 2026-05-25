// Package join merges multiple snapshot chunks into a single chunk by
// concatenating their payloads with a configurable delimiter.
package join

import (
	"errors"
	"io"
)

// Joiner accumulates chunks and emits a single merged chunk when Flush is
// called or the internal buffer reaches capacity.
type Joiner struct {
	delim     []byte
	maxChunks int
	buf       [][]byte
	onEmit    func([]byte) error
}

// New returns a Joiner that concatenates up to maxChunks payloads separated by
// delim, calling onEmit with the result on each flush.
//
// maxChunks must be >= 1 and onEmit must not be nil.
func New(delim []byte, maxChunks int, onEmit func([]byte) error) (*Joiner, error) {
	if maxChunks < 1 {
		return nil, errors.New("join: maxChunks must be >= 1")
	}
	if onEmit == nil {
		return nil, errors.New("join: onEmit must not be nil")
	}
	return &Joiner{
		delim:     delim,
		maxChunks: maxChunks,
		onEmit:    onEmit,
	}, nil
}

// Add appends p to the internal buffer. If the buffer reaches maxChunks the
// joiner is automatically flushed.
func (j *Joiner) Add(p []byte) error {
	if len(p) == 0 {
		return nil
	}
	chunk := make([]byte, len(p))
	copy(chunk, p)
	j.buf = append(j.buf, chunk)
	if len(j.buf) >= j.maxChunks {
		return j.Flush()
	}
	return nil
}

// Flush concatenates all buffered chunks with the delimiter and calls onEmit.
// It is a no-op when the buffer is empty.
func (j *Joiner) Flush() error {
	if len(j.buf) == 0 {
		return nil
	}
	var total int
	for _, c := range j.buf {
		total += len(c)
	}
	total += len(j.delim) * (len(j.buf) - 1)
	out := make([]byte, 0, total)
	for i, c := range j.buf {
		if i > 0 {
			out = append(out, j.delim...)
		}
		out = append(out, c...)
	}
	j.buf = j.buf[:0]
	return j.onEmit(out)
}

// Len returns the number of chunks currently buffered.
func (j *Joiner) Len() int { return len(j.buf) }

// NewWriter wraps dst so that every Write call feeds the joiner; the caller
// must Close the writer to flush any remaining buffered chunks.
func NewWriter(dst io.Writer, delim []byte, maxChunks int) (*Writer, error) {
	j, err := New(delim, maxChunks, func(p []byte) error {
		_, werr := dst.Write(p)
		return werr
	})
	if err != nil {
		return nil, err
	}
	return &Writer{j: j}, nil
}
