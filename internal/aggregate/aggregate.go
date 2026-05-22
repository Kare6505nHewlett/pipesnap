// Package aggregate batches consecutive chunks into fixed-size windows
// before forwarding them downstream as a single merged chunk.
package aggregate

import (
	"errors"
	"sync"
)

// Aggregator collects chunks and emits a merged chunk once the batch
// size is reached or Flush is called explicitly.
type Aggregator struct {
	mu      sync.Mutex
	size    int
	buf     [][]byte
	total   int
	onEmit  func([]byte) error
}

// New returns an Aggregator that batches up to size chunks before
// calling onEmit with the concatenated payload.
// size must be >= 1 and onEmit must not be nil.
func New(size int, onEmit func([]byte) error) (*Aggregator, error) {
	if size < 1 {
		return nil, errors.New("aggregate: size must be >= 1")
	}
	if onEmit == nil {
		return nil, errors.New("aggregate: onEmit must not be nil")
	}
	return &Aggregator{
		size:   size,
		buf:    make([][]byte, 0, size),
		onEmit: onEmit,
	}, nil
}

// Add appends p to the current batch. If the batch reaches the
// configured size it is flushed automatically.
func (a *Aggregator) Add(p []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	chunk := make([]byte, len(p))
	copy(chunk, p)
	a.buf = append(a.buf, chunk)
	a.total += len(p)

	if len(a.buf) >= a.size {
		return a.flush()
	}
	return nil
}

// Flush emits any buffered chunks immediately, even if the batch is
// not yet full. It is a no-op when the buffer is empty.
func (a *Aggregator) Flush() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.flush()
}

// Len returns the number of chunks currently buffered.
func (a *Aggregator) Len() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.buf)
}

func (a *Aggregator) flush() error {
	if len(a.buf) == 0 {
		return nil
	}
	merged := make([]byte, 0, a.total)
	for _, b := range a.buf {
		merged = append(merged, b...)
	}
	a.buf = a.buf[:0]
	a.total = 0
	return a.onEmit(merged)
}
