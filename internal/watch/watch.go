// Package watch provides a simple file-system watcher that emits an event
// whenever a snapshot file is modified or created. It is used by the replay
// command to automatically re-replay a snapshot when the source pipeline
// writes a new one.
package watch

import (
	"errors"
	"os"
	"time"
)

// Event is emitted on the channel returned by Watch.
type Event struct {
	Path    string
	ModTime time.Time
}

// Watcher polls a file path at the given interval and sends an Event on the
// returned channel whenever the file's modification time changes.
type Watcher struct {
	path     string
	interval time.Duration
	events   chan Event
	stop     chan struct{}
}

// New creates a new Watcher for the given path and poll interval.
// Call Start to begin watching.
func New(path string, interval time.Duration) *Watcher {
	return &Watcher{
		path:     path,
		interval: interval,
		events:   make(chan Event, 4),
		stop:     make(chan struct{}),
	}
}

// Events returns the read-only channel on which change events are delivered.
func (w *Watcher) Events() <-chan Event {
	return w.events
}

// Start begins polling in a background goroutine.
func (w *Watcher) Start() {
	go w.poll()
}

// Stop signals the background goroutine to exit and closes the events channel.
func (w *Watcher) Stop() {
	close(w.stop)
}

func (w *Watcher) poll() {
	defer close(w.events)

	var lastMod time.Time
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
			info, err := os.Stat(w.path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				continue
			}
			if mt := info.ModTime(); mt.After(lastMod) {
				lastMod = mt
				w.events <- Event{Path: w.path, ModTime: mt}
			}
		}
	}
}
