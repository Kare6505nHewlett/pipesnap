// Package pivot provides a chunk transformer that extracts a JSON field
// and re-emits chunks grouped under a new top-level key derived from that field.
package pivot

import (
	"encoding/json"
	"fmt"
	"io"
)

// Pivotter reads chunks whose payload is a JSON object, extracts the value
// at Field, and re-emits a new JSON object of the form:
//
//	{ "<value>": <original chunk> }
//
// If the field is missing or the chunk is not valid JSON the chunk is dropped
// and the drop count is incremented.
type Pivotter struct {
	field  string
	onDrop func([]byte)
	seen   int
	dropped int
}

// New creates a Pivotter that pivots on the given JSON field name.
// onDrop is called for every chunk that cannot be pivoted; it may be nil.
func New(field string, onDrop func([]byte)) (*Pivotter, error) {
	if field == "" {
		return nil, fmt.Errorf("pivot: field must not be empty")
	}
	return &Pivotter{field: field, onDrop: onDrop}, nil
}

// Apply transforms a single chunk. It returns the pivoted bytes and true on
// success, or nil and false when the chunk should be dropped.
func (p *Pivotter) Apply(chunk []byte) ([]byte, bool) {
	p.seen++

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(chunk, &obj); err != nil {
		p.drop(chunk)
		return nil, false
	}

	raw, ok := obj[p.field]
	if !ok {
		p.drop(chunk)
		return nil, false
	}

	// key is the string value of the pivot field
	var key string
	if err := json.Unmarshal(raw, &key); err != nil {
		p.drop(chunk)
		return nil, false
	}

	out, err := json.Marshal(map[string]json.RawMessage{key: chunk})
	if err != nil {
		p.drop(chunk)
		return nil, false
	}
	return out, true
}

// Seen returns the total number of chunks processed.
func (p *Pivotter) Seen() int { return p.seen }

// Dropped returns the number of chunks that were dropped.
func (p *Pivotter) Dropped() int { return p.dropped }

func (p *Pivotter) drop(chunk []byte) {
	p.dropped++
	if p.onDrop != nil {
		p.onDrop(chunk)
	}
}

// NewWriter wraps dst so that every Write call passes through the Pivotter.
func NewWriter(dst io.Writer, pv *Pivotter) io.WriteCloser {
	return &pivotWriter{dst: dst, pv: pv}
}

type pivotWriter struct {
	dst io.Writer
	pv  *Pivotter
}

func (w *pivotWriter) Write(p []byte) (int, error) {
	out, ok := w.pv.Apply(p)
	if !ok {
		return len(p), nil
	}
	if _, err := w.dst.Write(out); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *pivotWriter) Close() error {
	if c, ok := w.dst.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
