// Package sequence assigns monotonically increasing sequence numbers to
// chunks passing through a pipeline, enabling downstream consumers to detect
// gaps or reorder out-of-order delivery.
package sequence

import (
	"encoding/json"
	"fmt"
	"io"
	"sync/atomic"
)

// Sequencer wraps chunks with a sequence number injected into JSON payloads.
type Sequencer struct {
	counter uint64
	field   string
}

// New creates a Sequencer that injects seq numbers under the given field name.
// The counter starts at start (typically 0 or 1).
func New(field string, start uint64) (*Sequencer, error) {
	if field == "" {
		return nil, fmt.Errorf("sequence: field name must not be empty")
	}
	return &Sequencer{counter: start, field: field}, nil
}

// Next returns the next sequence number without applying it to any chunk.
func (s *Sequencer) Next() uint64 {
	return atomic.AddUint64(&s.counter, 1) - 1
}

// Apply injects the next sequence number into p, which must be a JSON object.
// The modified JSON is returned; p is never mutated.
func (s *Sequencer) Apply(p []byte) ([]byte, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(p, &obj); err != nil {
		return nil, fmt.Errorf("sequence: payload is not a JSON object: %w", err)
	}
	seq := atomic.AddUint64(&s.counter, 1) - 1
	obj[s.field] = seq
	out, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("sequence: marshal failed: %w", err)
	}
	return out, nil
}

// NewWriter wraps dst, applying sequence numbers to every chunk written.
// Chunks that are not valid JSON objects are dropped and an error is recorded
// but writing continues.
func NewWriter(dst io.Writer, s *Sequencer) *Writer {
	return &Writer{dst: dst, seq: s}
}

// Writer is an io.Writer that sequences each Write call.
type Writer struct {
	dst  io.Writer
	seq  *Sequencer
	Drops int
}

func (w *Writer) Write(p []byte) (int, error) {
	out, err := w.seq.Apply(p)
	if err != nil {
		w.Drops++
		return len(p), nil
	}
	if _, err := w.dst.Write(out); err != nil {
		return 0, err
	}
	return len(p), nil
}
