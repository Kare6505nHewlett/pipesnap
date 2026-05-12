package stats

import (
	"sync"
	"time"
)

// Collector tracks byte and chunk counts for a pipeline session.
type Collector struct {
	mu         sync.Mutex
	StartTime  time.Time
	Bytes      int64
	Chunks     int64
	Dropped    int64
}

// New creates a new Collector with the start time set to now.
func New() *Collector {
	return &Collector{
		StartTime: time.Now(),
	}
}

// RecordChunk records a successfully written chunk of the given size.
func (c *Collector) RecordChunk(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Bytes += int64(n)
	c.Chunks++
}

// RecordDrop records a dropped chunk of the given size.
func (c *Collector) RecordDrop(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Dropped++
	_ = n
}

// Elapsed returns the duration since the collector was started.
func (c *Collector) Elapsed() time.Duration {
	return time.Since(c.StartTime)
}

// Summary returns a point-in-time snapshot of the collected stats.
func (c *Collector) Summary() Summary {
	c.mu.Lock()
	defer c.mu.Unlock()
	return Summary{
		Elapsed: time.Since(c.StartTime),
		Bytes:   c.Bytes,
		Chunks:  c.Chunks,
		Dropped: c.Dropped,
	}
}

// Summary holds a read-only snapshot of collected statistics.
type Summary struct {
	Elapsed time.Duration
	Bytes   int64
	Chunks  int64
	Dropped int64
}
