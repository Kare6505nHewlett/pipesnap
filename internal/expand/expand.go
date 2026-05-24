// Package expand provides a chunk transformer that expands a single chunk
// into multiple chunks by splitting on a delimiter. This is useful when a
// snapshot chunk contains several logical records concatenated together and
// downstream stages expect one record per chunk.
package expand

import (
	"bytes"
	"errors"
	"io"
)

// Expander splits each chunk on Delim and emits one sub-chunk per segment.
// Empty segments produced by the split are silently dropped.
type Expander struct {
	delim []byte
}

// New returns an Expander that splits on delim.
// delim must be non-empty.
func New(delim []byte) (*Expander, error) {
	if len(delim) == 0 {
		return nil, errors.New("expand: delimiter must not be empty")
	}
	d := make([]byte, len(delim))
	copy(d, delim)
	return &Expander{delim: d}, nil
}

// Apply splits p on the configured delimiter and writes each non-empty segment
// to dst. It returns the length of p and any write error encountered.
func (e *Expander) Apply(dst io.Writer, p []byte) (int, error) {
	segments := bytes.Split(p, e.delim)
	for _, seg := range segments {
		if len(seg) == 0 {
			continue
		}
		if _, err := dst.Write(seg); err != nil {
			return len(p), err
		}
	}
	return len(p), nil
}

// NewWriter wraps dst with an Expander so that every Write call fans out the
// incoming bytes into multiple writes, one per delimiter-separated segment.
func NewWriter(dst io.Writer, delim []byte) (io.WriteCloser, error) {
	e, err := New(delim)
	if err != nil {
		return nil, err
	}
	return &expandWriter{dst: dst, exp: e}, nil
}

type expandWriter struct {
	dst io.Writer
	exp *Expander
}

func (w *expandWriter) Write(p []byte) (int, error) {
	return w.exp.Apply(w.dst, p)
}

func (w *expandWriter) Close() error {
	if c, ok := w.dst.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
