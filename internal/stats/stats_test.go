package stats

import (
	"testing"
	"time"
)

func TestNewCollectorDefaults(t *testing.T) {
	c := New()
	if c.Bytes != 0 || c.Chunks != 0 || c.Dropped != 0 {
		t.Fatal("expected zero values on new collector")
	}
	if c.StartTime.IsZero() {
		t.Fatal("expected start time to be set")
	}
}

func TestRecordChunk(t *testing.T) {
	c := New()
	c.RecordChunk(128)
	c.RecordChunk(256)
	if c.Bytes != 384 {
		t.Fatalf("expected 384 bytes, got %d", c.Bytes)
	}
	if c.Chunks != 2 {
		t.Fatalf("expected 2 chunks, got %d", c.Chunks)
	}
}

func TestRecordDrop(t *testing.T) {
	c := New()
	c.RecordDrop(64)
	c.RecordDrop(64)
	if c.Dropped != 2 {
		t.Fatalf("expected 2 dropped, got %d", c.Dropped)
	}
}

func TestElapsed(t *testing.T) {
	c := New()
	time.Sleep(10 * time.Millisecond)
	if c.Elapsed() < 10*time.Millisecond {
		t.Fatal("expected elapsed to be at least 10ms")
	}
}

func TestSummary(t *testing.T) {
	c := New()
	c.RecordChunk(100)
	c.RecordDrop(50)
	s := c.Summary()
	if s.Bytes != 100 {
		t.Fatalf("expected 100 bytes in summary, got %d", s.Bytes)
	}
	if s.Chunks != 1 {
		t.Fatalf("expected 1 chunk in summary, got %d", s.Chunks)
	}
	if s.Dropped != 1 {
		t.Fatalf("expected 1 dropped in summary, got %d", s.Dropped)
	}
	if s.Elapsed <= 0 {
		t.Fatal("expected positive elapsed in summary")
	}
}
