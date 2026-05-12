package ratelimit_test

import (
	"testing"
	"time"

	"pipesnap/internal/ratelimit"
)

func TestNoOpLimiter(t *testing.T) {
	l := ratelimit.New(ratelimit.Config{})
	if !l.IsNoOp() {
		t.Fatal("expected no-op limiter")
	}
	// Wait should return instantly
	start := time.Now()
	l.Wait(1024)
	if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
		t.Fatalf("no-op limiter waited too long: %v", elapsed)
	}
}

func TestMinDelayIsRespected(t *testing.T) {
	delay := 50 * time.Millisecond
	l := ratelimit.New(ratelimit.Config{MinDelay: delay})
	if l.IsNoOp() {
		t.Fatal("expected non-no-op limiter")
	}
	start := time.Now()
	l.Wait(0)
	elapsed := time.Since(start)
	if elapsed < delay {
		t.Fatalf("expected at least %v delay, got %v", delay, elapsed)
	}
}

func TestBytesPerSecThrottle(t *testing.T) {
	// 1000 bytes/s → 100-byte chunk should take ~100 ms
	l := ratelimit.New(ratelimit.Config{BytesPerSec: 1000})
	start := time.Now()
	l.Wait(100)
	elapsed := time.Since(start)
	expected := 100 * time.Millisecond
	if elapsed < expected-10*time.Millisecond {
		t.Fatalf("throttle too fast: got %v, want >= %v", elapsed, expected)
	}
}

func TestMinDelayWinsOverThroughput(t *testing.T) {
	// minDelay > throughput delay → minDelay should win
	minDelay := 80 * time.Millisecond
	// 10000 bytes/s → 10-byte chunk ≈ 1 ms throttle
	l := ratelimit.New(ratelimit.Config{BytesPerSec: 10000, MinDelay: minDelay})
	start := time.Now()
	l.Wait(10)
	elapsed := time.Since(start)
	if elapsed < minDelay-10*time.Millisecond {
		t.Fatalf("expected minDelay to dominate; got %v", elapsed)
	}
}

func TestThroughputWinsOverMinDelay(t *testing.T) {
	// throughput delay > minDelay → throughput should win
	// 100 bytes/s → 100-byte chunk ≈ 1000 ms; cap test to something small
	// Use 2000 bytes/s → 100-byte chunk ≈ 50 ms; minDelay = 10 ms
	l := ratelimit.New(ratelimit.Config{BytesPerSec: 2000, MinDelay: 10 * time.Millisecond})
	start := time.Now()
	l.Wait(100)
	elapsed := time.Since(start)
	expected := 50 * time.Millisecond
	if elapsed < expected-15*time.Millisecond {
		t.Fatalf("expected throughput to dominate; got %v", elapsed)
	}
}
