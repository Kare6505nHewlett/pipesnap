// Package truncate provides helpers for trimming snapshot files down to a
// bounded size by discarding the oldest chunks.
//
// Usage
//
// Given a snapshot produced by the snapshot package you can enforce retention
// limits with a single call:
//
//	err := truncate.File("/var/snap/data.bin", "/var/snap/data.bin", truncate.Options{
//		MaxChunks: 1000,
//		MaxBytes:  10 * 1024 * 1024, // 10 MiB
//	})
//
// When both MaxChunks and MaxBytes are set the stricter constraint wins: the
// chunk-count limit is applied first, then the byte limit is applied to the
// surviving chunks.
//
// If both limits are zero the function returns immediately without touching
// the file.
//
// The write is atomic: the result is first written to a sibling temp file and
// then renamed into place, so a crash mid-write will not corrupt the original.
package truncate
