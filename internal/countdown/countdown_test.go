package countdown_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/your-org/pipesnap/internal/countdown"
)

// nopCloser wraps a *bytes.Buffer so it satisfies io.WriteCloser.
type nopCloser struct {
	*bytes.Buffer
	closed bool
}

func (n *nopCloser) Close() error { n.closed = true; return nil }

func newDst() *nopCloser { return &nopCloser{Buffer: &bytes.Buffer{}} }

func TestNewRejectsNilDst(t *testing.T) {
	_, err := countdown.New(nil, 3)
	if err == nil {
		t.Fatal("expected error for nil dst")
	}
}

func TestNewRejectsZeroN(t *testing.T) {
	_, err := countdown.New(newDst(), 0)
	if err == nil {
		t.Fatal("expected error for n=0")
	}
}

func TestNewRejectsNegativeN(t *testing.T) {
	_, err := countdown.New(newDst(), -1)
	if err == nil {
		t.Fatal("expected error for n<0")
	}
}

func TestWriteForwardsChunks(t *testing.T) {
	dst := newDst()
	c, _ := countdown.New(dst, 3)

	for i := 0; i < 3; i++ {
		_, err := c.Write([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error on write %d: %v", i, err)
		}
	}

	if got := dst.String(); got != "hellohellohello" {
		t.Fatalf("unexpected buffer content: %q", got)
	}
}

func TestWriteReturnErrLimitReachedAfterN(t *testing.T) {
	dst := newDst()
	c, _ := countdown.New(dst, 2)

	c.Write([]byte("a")) // nolint
	c.Write([]byte("b")) // nolint

	_, err := c.Write([]byte("c"))
	if !errors.Is(err, countdown.ErrLimitReached) {
		t.Fatalf("expected ErrLimitReached, got %v", err)
	}
}

func TestDstClosedWhenLimitReached(t *testing.T) {
	dst := newDst()
	c, _ := countdown.New(dst, 1)

	c.Write([]byte("x")) // nolint

	if !dst.closed {
		t.Fatal("expected dst to be closed after limit")
	}
	if !c.Done() {
		t.Fatal("expected Done() to return true")
	}
}

func TestRemainingDecrementsOnWrite(t *testing.T) {
	dst := newDst()
	c, _ := countdown.New(dst, 3)

	if c.Remaining() != 3 {
		t.Fatalf("expected 3, got %d", c.Remaining())
	}
	c.Write([]byte("a")) // nolint
	if c.Remaining() != 2 {
		t.Fatalf("expected 2, got %d", c.Remaining())
	}
}

func TestCloseBeforeLimitClosesDst(t *testing.T) {
	dst := newDst()
	c, _ := countdown.New(dst, 5)

	if err := c.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if !dst.closed {
		t.Fatal("expected dst to be closed")
	}
}

func TestCloseAfterLimitIsNoOp(t *testing.T) {
	dst := newDst()
	c, _ := countdown.New(dst, 1)
	c.Write([]byte("x")) // nolint — triggers close

	// Second close should not error even though dst is already closed.
	if err := c.Close(); err != nil {
		t.Fatalf("unexpected error on second close: %v", err)
	}
}

// Ensure Countdown satisfies io.WriteCloser.
var _ io.WriteCloser = (*countdown.Countdown)(nil)
