// Package gate implements a conditional pass-through stage for pipesnap
// pipelines.
//
// A Gate starts in the closed state and silently drops every chunk until a
// caller-supplied predicate returns true for an incoming chunk. From that
// point on the gate stays open and all subsequent chunks are forwarded to the
// destination writer.
//
// The gate can also be forced open or closed at any time via Open() and
// Close(), making it useful for controlling stream flow based on external
// signals such as a checkpoint being reached or a watch event firing.
//
// Typical usage:
//
//	g, err := gate.New(dst, func(p []byte) bool {
//		return bytes.Contains(p, []byte("START"))
//	})
//	// pipe chunks through g.Write — forwarding begins on the first
//	// chunk that contains "START".
package gate
