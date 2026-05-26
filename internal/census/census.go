// Package census counts the frequency of distinct chunk values seen
// in a stream. It is useful for debugging pipelines where you want to
// understand the distribution of data flowing through a stage.
package census

import (
	"fmt"
	"io"
	"sort"
	"sync"
)

// Entry holds a distinct value and the number of times it was seen.
type Entry struct {
	Value []byte
	Count int64
}

// Census tracks value frequencies for chunks written to it.
type Census struct {
	mu      sync.Mutex
	counts  map[string]int64
	total   int64
	dropped int64
	maxKeys int
}

// New returns a new Census. maxKeys limits the number of distinct values
// tracked; once the limit is reached additional unseen values are counted
// as dropped. Pass 0 for unlimited.
func New(maxKeys int) (*Census, error) {
	if maxKeys < 0 {
		return nil, fmt.Errorf("census: maxKeys must be >= 0, got %d", maxKeys)
	}
	return &Census{
		counts:  make(map[string]int64),
		maxKeys: maxKeys,
	}, nil
}

// Record registers a single chunk value.
func (c *Census) Record(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.total++
	key := string(data)
	if _, exists := c.counts[key]; !exists {
		if c.maxKeys > 0 && len(c.counts) >= c.maxKeys {
			c.dropped++
			return
		}
		c.counts[key] = 0
	}
	c.counts[key]++
}

// Entries returns all tracked entries sorted by descending count.
func (c *Census) Entries() []Entry {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Entry, 0, len(c.counts))
	for k, v := range c.counts {
		out = append(out, Entry{Value: []byte(k), Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return string(out[i].Value) < string(out[j].Value)
	})
	return out
}

// Total returns the total number of chunks recorded (including dropped).
func (c *Census) Total() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.total
}

// Dropped returns the number of chunks not tracked due to the key limit.
func (c *Census) Dropped() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.dropped
}

// NewWriter returns an io.WriteCloser that records every written chunk in c
// and forwards it to dst.
func NewWriter(dst io.WriteCloser, c *Census) io.WriteCloser {
	return &censusWriter{dst: dst, census: c}
}

type censusWriter struct {
	dst    io.WriteCloser
	census *Census
}

func (w *censusWriter) Write(p []byte) (int, error) {
	w.census.Record(p)
	return w.dst.Write(p)
}

func (w *censusWriter) Close() error {
	return w.dst.Close()
}
