package watch_test

import (
	"os"
	"testing"
	"time"

	"github.com/user/pipesnap/internal/watch"
)

func TestEventsDeliveredOnChange(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "snap-*.bin")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	w := watch.New(f.Name(), 20*time.Millisecond)
	w.Start()
	defer w.Stop()

	// Give the watcher one tick to record the initial mod-time.
	time.Sleep(40 * time.Millisecond)

	// Touch the file.
	if err := os.Chtimes(f.Name(), time.Now(), time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	select {
	case ev := <-w.Events():
		if ev.Path != f.Name() {
			t.Fatalf("expected path %q, got %q", f.Name(), ev.Path)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for event")
	}
}

func TestNoEventForMissingFile(t *testing.T) {
	w := watch.New("/tmp/pipesnap-does-not-exist-xyz.bin", 20*time.Millisecond)
	w.Start()
	defer w.Stop()

	select {
	case ev, ok := <-w.Events():
		if ok {
			t.Fatalf("unexpected event for missing file: %+v", ev)
		}
		// channel closed after Stop — ok
	case <-time.After(120 * time.Millisecond):
		// expected: no events
	}
}

func TestStopClosesChannel(t *testing.T) {
	w := watch.New("/tmp/pipesnap-stop-test.bin", 10*time.Millisecond)
	w.Start()
	w.Stop()

	// After Stop, the channel must eventually be closed.
	deadline := time.After(200 * time.Millisecond)
	for {
		select {
		case _, ok := <-w.Events():
			if !ok {
				return // closed as expected
			}
		case <-deadline:
			t.Fatal("events channel was not closed after Stop")
		}
	}
}

func TestEventContainsModTime(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "snap-*.bin")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	w := watch.New(f.Name(), 20*time.Millisecond)
	w.Start()
	defer w.Stop()

	time.Sleep(40 * time.Millisecond)

	want := time.Now().Add(2 * time.Second).Truncate(time.Second)
	if err := os.Chtimes(f.Name(), want, want); err != nil {
		t.Fatal(err)
	}

	select {
	case ev := <-w.Events():
		if !ev.ModTime.Equal(want) && ev.ModTime.Before(want) {
			t.Fatalf("expected modtime >= %v, got %v", want, ev.ModTime)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for event")
	}
}
