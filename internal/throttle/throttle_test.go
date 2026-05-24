package throttle

import (
	"testing"
	"time"
)

func TestNewRejectsZeroMax(t *testing.T) {
	_, err := New(0, time.Second)
	if err == nil {
		t.Fatal("expected error for zero maxPerWindow")
	}
}

func TestNewRejectsNegativeWindow(t *testing.T) {
	_, err := New(5, -time.Second)
	if err == nil {
		t.Fatal("expected error for negative window")
	}
}

func TestAllowUpToMax(t *testing.T) {
	th, err := New(3, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if !th.Allow() {
			t.Fatalf("expected Allow()=true on call %d", i+1)
		}
	}
	if th.Allow() {
		t.Fatal("expected Allow()=false after budget exhausted")
	}
}

func TestRemainingDecrements(t *testing.T) {
	th, _ := New(4, time.Hour)
	if th.Remaining() != 4 {
		t.Fatalf("want 4, got %d", th.Remaining())
	}
	th.Allow()
	if th.Remaining() != 3 {
		t.Fatalf("want 3, got %d", th.Remaining())
	}
}

func TestWindowReset(t *testing.T) {
	th, err := New(2, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	th.Allow()
	th.Allow()
	if th.Allow() {
		t.Fatal("budget should be exhausted")
	}
	time.Sleep(30 * time.Millisecond)
	if !th.Allow() {
		t.Fatal("expected Allow()=true after window reset")
	}
}

func TestRemainingAfterWindowReset(t *testing.T) {
	th, _ := New(3, 20*time.Millisecond)
	th.Allow()
	th.Allow()
	th.Allow()
	time.Sleep(30 * time.Millisecond)
	if got := th.Remaining(); got != 3 {
		t.Fatalf("want 3 after reset, got %d", got)
	}
}
