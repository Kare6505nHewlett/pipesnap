// Package dedupe implements content-based deduplication for pipesnap
// snapshot streams.
//
// A Filter tracks the SHA-256 hash of every chunk it has processed. When
// a chunk arrives whose hash is already in the set, IsDuplicate returns
// true and the caller can skip writing that chunk.
//
// NewWriter wraps any io.Writer and transparently drops duplicate chunks
// so downstream writers never see repeated content. The underlying
// writer receives each unique chunk exactly once.
//
// Example:
//
//	f := dedupe.New()
//	w := dedupe.NewWriter(snapshotWriter, f)
//	// pipe stdin through w — identical lines are silently dropped
//	io.Copy(w, os.Stdin)
package dedupe
