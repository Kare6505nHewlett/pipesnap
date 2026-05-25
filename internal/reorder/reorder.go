// Package reorder buffers chunks and emits them in a deterministic order
// based on a numeric sequence field embedded in each JSON chunk.
// Chunks arriving out of order are held in a priority queue until
// the expected next sequence number is available.
package reorder

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"io"
)

// item holds a buffered chunk with its parsed sequence number.
type item struct {
	seq  int64
	data []byte
}

type minHeap []item

func (h minHeap) Len() int            { return len(h) }
func (h minHeap) Less(i, j int) bool  { return h[i].seq < h[j].seq }
func (h minHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) { *h = append(*h, x.(item)) }
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// Reorder buffers out-of-order chunks and emits them in sequence order.
type Reorder struct {
	field  string
	next   int64
	buf    minHeap
	onEmit func([]byte) error
}

// New creates a Reorder that reads seq numbers from field and calls onEmit
// for each chunk in order. startSeq is the first expected sequence number.
func New(field string, startSeq int64, onEmit func([]byte) error) (*Reorder, error) {
	if field == "" {
		return nil, fmt.Errorf("reorder: field must not be empty")
	}
	if onEmit == nil {
		return nil, fmt.Errorf("reorder: onEmit must not be nil")
	}
	h := minHeap{}
	heap.Init(&h)
	return &Reorder{field: field, next: startSeq, buf: h, onEmit: onEmit}, nil
}

// Add buffers chunk and flushes any consecutive sequence run to onEmit.
func (r *Reorder) Add(chunk []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(chunk, &m); err != nil {
		return fmt.Errorf("reorder: invalid JSON: %w", err)
	}
	v, ok := m[r.field]
	if !ok {
		return fmt.Errorf("reorder: field %q not found", r.field)
	}
	seqF, ok := v.(float64)
	if !ok {
		return fmt.Errorf("reorder: field %q is not a number", r.field)
	}
	heap.Push(&r.buf, item{seq: int64(seqF), data: chunk})
	return r.flush()
}

// Flush emits all remaining buffered chunks regardless of gaps.
func (r *Reorder) Flush() error {
	for r.buf.Len() > 0 {
		it := heap.Pop(&r.buf).(item)
		if err := r.onEmit(it.data); err != nil {
			return err
		}
	}
	return nil
}

// Buffered returns the number of chunks currently held in the buffer.
func (r *Reorder) Buffered() int { return r.buf.Len() }

func (r *Reorder) flush() error {
	for r.buf.Len() > 0 && r.buf[0].seq == r.next {
		it := heap.Pop(&r.buf).(item)
		if err := r.onEmit(it.data); err != nil {
			return err
		}
		r.next++
	}
	return nil
}

// NewWriter wraps dst and reorders chunks before writing.
func NewWriter(dst io.Writer, field string, startSeq int64) (*Writer, error) {
	r, err := New(field, startSeq, func(chunk []byte) error {
		_, werr := dst.Write(chunk)
		return werr
	})
	if err != nil {
		return nil, err
	}
	return &Writer{r: r, dst: dst}, nil
}
