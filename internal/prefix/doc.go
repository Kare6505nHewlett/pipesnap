// Package prefix implements a chunk transform that prepends a fixed string to
// every chunk passing through a pipeline stage.
//
// It is useful for tagging stream data with a source label, log level, or any
// other constant marker before the data is written to a snapshot file or
// forwarded downstream.
//
// Basic usage:
//
//	pf, err := prefix.New("[INFO] ")
//	if err != nil { ... }
//	out, err := pf.Apply(chunk)
//
// Writer usage (wraps any io.Writer):
//
//	w, err := prefix.NewWriter(dst, "[INFO] ")
//	if err != nil { ... }
//	w.Write(chunk)
//	w.Close()
//
// Empty chunks are passed through without modification so that upstream writers
// that emit keep-alive zero-byte writes are not affected.
package prefix
