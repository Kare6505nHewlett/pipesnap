// Package audit provides a lightweight structured-logging layer for the
// pipesnap pipeline. Every chunk that passes through an audit.Writer is
// recorded as a newline-delimited JSON entry containing:
//
//   - at        – UTC timestamp of the observation
//   - label     – optional human-readable tag (e.g. "ingest", "post-filter")
//   - bytes     – exact byte length of the chunk
//   - preview   – first N bytes of the chunk body (default 64)
//   - truncated – true when the chunk body exceeded the preview limit
//
// Audit entries are written best-effort; a failure to write an entry never
// interrupts data flow.
//
// Typical usage:
//
//	logger, _ := audit.New(os.Stderr, "ingest", 0)
//	w, _      := audit.NewWriter(snapshotWriter, logger)
//	io.Copy(w, os.Stdin)
package audit
