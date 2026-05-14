// Package window provides a sliding window buffer that retains the last N
// chunks written to it, useful for capturing recent history from a stream.
package window

import "sync"

// Chunk holds a single captured data segment.
type Chunk struct {
	Data []byte
}

// Buffer is a fixed-capacity circular buffer of chunks.
type Buffer struct {
	mu       sync.Mutex
	capacity int
	chunks   []Chunk
	head     int // index of the oldest entry
	count    int
}

// New creates a Buffer that retains at most capacity chunks.
// If capacity is <= 0 it defaults to 1.
func New(capacity int) *Buffer {
	if capacity <= 0 {
		capacity = 1
	}
	return &Buffer{
		capacity: capacity,
		chunks:   make([]Chunk, capacity),
	}
}

// Write appends a chunk to the buffer, evicting the oldest entry when full.
// The data slice is copied so the caller may reuse it safely.
func (b *Buffer) Write(data []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	copy := make([]byte, len(data))
	copy = append(copy[:0], data...)

	if b.count < b.capacity {
		// Still filling up: write at head+count position.
		idx := (b.head + b.count) % b.capacity
		b.chunks[idx] = Chunk{Data: copy}
		b.count++
	} else {
		// Full: overwrite oldest slot and advance head.
		b.chunks[b.head] = Chunk{Data: copy}
		b.head = (b.head + 1) % b.capacity
	}
}

// Snapshot returns an ordered slice of all retained chunks, oldest first.
func (b *Buffer) Snapshot() []Chunk {
	b.mu.Lock()
	defer b.mu.Unlock()

	out := make([]Chunk, b.count)
	for i := 0; i < b.count; i++ {
		idx := (b.head + i) % b.capacity
		dst := make([]byte, len(b.chunks[idx].Data))
		copy(dst, b.chunks[idx].Data)
		out[i] = Chunk{Data: dst}
	}
	return out
}

// Len returns the number of chunks currently held.
func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.count
}

// Reset clears all retained chunks.
func (b *Buffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.head = 0
	b.count = 0
}
