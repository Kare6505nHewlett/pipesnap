// Package expand provides a chunk transformer that splits a single incoming
// chunk into multiple smaller chunks using a configurable byte delimiter.
//
// # Overview
//
// Some pipeline stages emit concatenated records in a single write — for
// example newline-delimited JSON blobs or pipe-separated values.  The expand
// package re-splits those writes so that each downstream stage receives one
// logical record per chunk, which is the convention assumed by most other
// pipesnap packages.
//
// # Usage
//
//	e, err := expand.New([]byte("\n"))
//	if err != nil { ... }
//	n, err := e.Apply(dst, chunk)
//
// Or use the convenience writer wrapper:
//
//	w, err := expand.NewWriter(dst, []byte("\n"))
//	if err != nil { ... }
//	defer w.Close()
//	io.Copy(w, src)
//
// Empty segments produced by consecutive delimiters are silently dropped.
package expand
