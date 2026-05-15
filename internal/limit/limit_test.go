package limit_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/yourorg/pipesnap/internal/limit"
)

func TestAllowUnlimited(t *testing.T) {
	l := limit.New(0, 0)
	for i := 0; i < 1000; i++ {
		if !l.Allow(512) {
			t.Fatalf("expected Allow to return true for unlimited limiter at iteration %d", i)
		}
	}
}

func TestAllowByChunks(t *testing.T) {
	l := limit.New(3, 0)
	for i := 0; i < 3; i++ {
		if !l.Allow(1) {
			t.Fatalf("chunk %d should be allowed", i)
		}
	}
	if l.Allow(1) {
		t.Fatal("fourth chunk should be denied")
	}
}

func TestAllowByBytes(t *testing.T) {
	l := limit.New(0, 10)
	if !l.Allow(7) {
		t.Fatal("7 bytes should be allowed")
	}
	// next chunk would push total to 14 > 10
	if l.Allow(7) {
		t.Fatal("second 7-byte chunk should be denied")
	}
}

func TestCounters(t *testing.T) {
	l := limit.New(5, 100)
	l.Allow(20)
	l.Allow(30)
	if l.Chunks() != 2 {
		t.Fatalf("expected 2 chunks, got %d", l.Chunks())
	}
	if l.Bytes() != 50 {
		t.Fatalf("expected 50 bytes, got %d", l.Bytes())
	}
}

func TestWrapReaderByChunks(t *testing.T) {
	data := bytes.Repeat([]byte("hello"), 10)
	r := bytes.NewReader(data)
	l := limit.New(2, 0)
	lr := limit.WrapReader(r, l)

	buf := make([]byte, 5)
	n, err := lr.Read(buf)
	if err != nil || n != 5 {
		t.Fatalf("first read: got n=%d err=%v", n, err)
	}
	n, err = lr.Read(buf)
	if err != nil || n != 5 {
		t.Fatalf("second read: got n=%d err=%v", n, err)
	}
	// third read should be blocked
	n, err = lr.Read(buf)
	if err != io.EOF {
		t.Fatalf("expected io.EOF after limit, got n=%d err=%v", n, err)
	}
}

func TestWrapReaderByBytes(t *testing.T) {
	data := []byte("abcdefghijklmnopqrstuvwxyz")
	r := bytes.NewReader(data)
	l := limit.New(0, 8)
	lr := limit.WrapReader(r, l)

	buf := make([]byte, 8)
	n, err := lr.Read(buf)
	if err != nil || n != 8 {
		t.Fatalf("first read: got n=%d err=%v", n, err)
	}
	n, err = lr.Read(buf)
	if err != io.EOF {
		t.Fatalf("expected io.EOF, got n=%d err=%v", n, err)
	}
}
