// Package tag provides a transform that injects a fixed string tag into
// every JSON chunk under a configurable field name. Non-JSON chunks are
// passed through unchanged.
package tag

import (
	"encoding/json"
	"fmt"
	"io"
)

// Tagger injects a fixed tag value into JSON chunks.
type Tagger struct {
	field string
	value string
}

// New returns a Tagger that sets chunk[field] = value.
// field must be non-empty; value may be any string.
func New(field, value string) (*Tagger, error) {
	if field == "" {
		return nil, fmt.Errorf("tag: field must not be empty")
	}
	return &Tagger{field: field, value: value}, nil
}

// Apply injects the tag into p if it is valid JSON, otherwise returns p
// unmodified. A non-nil error is returned only on internal marshal failures.
func (t *Tagger) Apply(p []byte) ([]byte, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(p, &obj); err != nil {
		// Not JSON — pass through.
		return p, nil
	}
	obj[t.field] = t.value
	out, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("tag: marshal: %w", err)
	}
	return out, nil
}

// NewWriter wraps dst so that every chunk written is tagged before being
// forwarded. Writes that fail Apply are dropped silently (the original
// length is still returned to satisfy io.Writer callers).
func NewWriter(dst io.Writer, field, value string) (io.Writer, error) {
	t, err := New(field, value)
	if err != nil {
		return nil, err
	}
	return &tagWriter{dst: dst, tagger: t}, nil
}

type tagWriter struct {
	dst    io.Writer
	tagger *Tagger
}

func (w *tagWriter) Write(p []byte) (int, error) {
	out, err := w.tagger.Apply(p)
	if err != nil {
		// Drop the chunk; report success so the caller keeps going.
		return len(p), nil
	}
	if _, err := w.dst.Write(out); err != nil {
		return 0, err
	}
	return len(p), nil
}
