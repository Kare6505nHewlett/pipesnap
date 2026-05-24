// Package batch groups consecutive chunks by a shared JSON field value,
// emitting a combined payload whenever the group key changes or the
// maximum batch size is reached.
package batch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Batcher groups chunks by the value of a JSON field.
type Batcher struct {
	field   string
	maxSize int
	onEmit  func(key string, chunks [][]byte) error

	currentKey string
	buf        [][]byte
}

// New creates a Batcher that groups by field and calls onEmit when a group
// is complete. maxSize is the maximum number of chunks per group (>= 1).
func New(field string, maxSize int, onEmit func(key string, chunks [][]byte) error) (*Batcher, error) {
	if field == "" {
		return nil, errors.New("batch: field must not be empty")
	}
	if maxSize < 1 {
		return nil, errors.New("batch: maxSize must be >= 1")
	}
	if onEmit == nil {
		return nil, errors.New("batch: onEmit must not be nil")
	}
	return &Batcher{field: field, maxSize: maxSize, onEmit: onEmit}, nil
}

// Add processes a single chunk. If the chunk's key differs from the current
// group, or the buffer is full, the current group is flushed first.
func (b *Batcher) Add(chunk []byte) error {
	key, err := b.extractKey(chunk)
	if err != nil {
		return fmt.Errorf("batch: %w", err)
	}

	if key != b.currentKey && len(b.buf) > 0 {
		if err := b.flush(); err != nil {
			return err
		}
	}

	b.currentKey = key
	b.buf = append(b.buf, chunk)

	if len(b.buf) >= b.maxSize {
		return b.flush()
	}
	return nil
}

// Flush emits any buffered chunks regardless of group size.
func (b *Batcher) Flush() error {
	if len(b.buf) == 0 {
		return nil
	}
	return b.flush()
}

// WriteTo writes all buffered chunks to w and resets the buffer.
func (b *Batcher) WriteTo(w io.Writer) (int64, error) {
	var total int64
	for _, chunk := range b.buf {
		n, err := w.Write(chunk)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}
	b.buf = b.buf[:0]
	return total, nil
}

func (b *Batcher) flush() error {
	copy := make([][]byte, len(b.buf))
	for i, c := range b.buf {
		copy[i] = c
	}
	b.buf = b.buf[:0]
	return b.onEmit(b.currentKey, copy)
}

func (b *Batcher) extractKey(chunk []byte) (string, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(chunk, &m); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	v, ok := m[b.field]
	if !ok {
		return "", nil
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v), nil
	}
	return s, nil
}
