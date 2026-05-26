// Package countdown provides a writer that forwards chunks until a
// fixed count is reached, then drops all subsequent chunks and closes
// the underlying destination.
package countdown

import (
	"errors"
	"io"
	"sync"
)

// ErrLimitReached is returned by Write once the chunk limit has been hit.
var ErrLimitReached = errors.New("countdown: limit reached")

// Countdown forwards the first N chunks to dst, then closes dst and
// returns ErrLimitReached for every subsequent Write call.
type Countdown struct {
	mu      sync.Mutex
	dst     io.WriteCloser
	remain  int
	done    bool
}

// New creates a Countdown that will forward exactly n chunks.
// n must be greater than zero.
func New(dst io.WriteCloser, n int) (*Countdown, error) {
	if dst == nil {
		return nil, errors.New("countdown: dst must not be nil")
	}
	if n <= 0 {
		return nil, errors.New("countdown: n must be greater than zero")
	}
	return &Countdown{dst: dst, remain: n}, nil
}

// Write forwards p to the underlying destination if the limit has not
// been reached. On the write that exhausts the limit the data is still
// forwarded and dst is closed. Subsequent writes return ErrLimitReached.
func (c *Countdown) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.done {
		return 0, ErrLimitReached
	}

	n, err := c.dst.Write(p)
	if err != nil {
		return n, err
	}

	c.remain--
	if c.remain <= 0 {
		c.done = true
		_ = c.dst.Close()
	}

	return n, nil
}

// Remaining returns the number of chunks that can still be forwarded.
func (c *Countdown) Remaining() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.remain
}

// Done reports whether the limit has been reached.
func (c *Countdown) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.done
}

// Close closes the underlying destination if it has not already been
// closed by the limit being reached.
func (c *Countdown) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.done {
		return nil
	}
	c.done = true
	return c.dst.Close()
}
