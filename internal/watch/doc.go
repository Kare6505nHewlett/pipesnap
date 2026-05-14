// Package watch implements a lightweight polling-based file watcher used by
// pipesnap to detect when a snapshot file has been updated on disk.
//
// # Overview
//
// A [Watcher] polls a single file path at a configurable interval. When the
// file's modification time advances, an [Event] is sent on the channel
// returned by [Watcher.Events]. The caller controls the lifecycle with
// [Watcher.Start] and [Watcher.Stop].
//
// # Usage
//
//	w := watch.New("/tmp/my.snap", 500*time.Millisecond)
//	w.Start()
//	defer w.Stop()
//
//	for ev := range w.Events() {
//		fmt.Println("snapshot updated:", ev.Path, ev.ModTime)
//	}
//
// # Design notes
//
// Polling is intentionally used instead of inotify/kqueue so that the package
// remains portable and dependency-free. For the typical pipesnap use-case
// (debugging pipelines that flush every few seconds) polling at ~500 ms
// introduces negligible latency while keeping the implementation trivial.
package watch
