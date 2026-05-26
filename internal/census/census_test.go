package census

import (
	"bytes"
	"io"
	"testing"
)

func TestNewRejectsNegativeMaxKeys(t *testing.T) {
	_, err := New(-1)
	if err == nil {
		t.Fatal("expected error for negative maxKeys")
	}
}

func TestNewAcceptsZeroMaxKeys(t *testing.T) {
	c, err := New(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil census")
	}
}

func TestRecordCountsValues(t *testing.T) {
	c, _ := New(0)
	c.Record([]byte("hello"))
	c.Record([]byte("hello"))
	c.Record([]byte("world"))

	if got := c.Total(); got != 3 {
		t.Fatalf("total: want 3, got %d", got)
	}
	entries := c.Entries()
	if len(entries) != 2 {
		t.Fatalf("entries: want 2, got %d", len(entries))
	}
	if string(entries[0].Value) != "hello" || entries[0].Count != 2 {
		t.Fatalf("top entry wrong: %+v", entries[0])
	}
	if string(entries[1].Value) != "world" || entries[1].Count != 1 {
		t.Fatalf("second entry wrong: %+v", entries[1])
	}
}

func TestEntriesSortedByCount(t *testing.T) {
	c, _ := New(0)
	for i := 0; i < 5; i++ {
		c.Record([]byte("a"))
	}
	for i := 0; i < 3; i++ {
		c.Record([]byte("b"))
	}
	c.Record([]byte("c"))

	entries := c.Entries()
	if entries[0].Count < entries[1].Count || entries[1].Count < entries[2].Count {
		t.Fatal("entries not in descending order")
	}
}

func TestDroppedWhenMaxKeysExceeded(t *testing.T) {
	c, _ := New(2)
	c.Record([]byte("a"))
	c.Record([]byte("b"))
	c.Record([]byte("c")) // should be dropped
	c.Record([]byte("a")) // existing key — should not be dropped

	if got := c.Dropped(); got != 1 {
		t.Fatalf("dropped: want 1, got %d", got)
	}
	if got := c.Total(); got != 4 {
		t.Fatalf("total: want 4, got %d", got)
	}
	if len(c.Entries()) != 2 {
		t.Fatalf("entries: want 2, got %d", len(c.Entries()))
	}
}

func TestNewWriterForwardsAndRecords(t *testing.T) {
	c, _ := New(0)
	var buf closeBuf
	w := NewWriter(&buf, c)

	_, _ = w.Write([]byte("ping"))
	_, _ = w.Write([]byte("ping"))
	_, _ = w.Write([]byte("pong"))
	_ = w.Close()

	if !buf.closed {
		t.Fatal("expected underlying writer to be closed")
	}
	if !bytes.Equal(buf.Bytes()[0:4], []byte("ping")) {
		t.Fatal("data not forwarded")
	}
	if got := c.Total(); got != 3 {
		t.Fatalf("total: want 3, got %d", got)
	}
}

// closeBuf is a bytes.Buffer that also tracks Close calls.
type closeBuf struct {
	bytes.Buffer
	closed bool
}

func (b *closeBuf) Close() error {
	b.closed = true
	return nil
}

var _ io.WriteCloser = (*closeBuf)(nil)
