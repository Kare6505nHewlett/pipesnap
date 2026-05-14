// Package rotate implements snapshot file rotation for pipesnap.
//
// It supports two rotation policies that can be combined:
//
//   - MaxFiles: keeps at most N snapshot files, removing the oldest first.
//   - MaxBytes: keeps total snapshot size under a byte threshold, removing
//     the oldest files first until the limit is satisfied.
//
// Usage:
//
//	m := rotate.New(rotate.Config{
//		Dir:      "/var/snapshots",
//		Prefix:   "myapp-",
//		MaxFiles: 10,
//		MaxBytes: 100 * 1024 * 1024, // 100 MiB
//	})
//
//	w, err := rotate.NewWriter(m)
//	if err != nil { ... }
//	defer w.Close()
//	io.Copy(w, os.Stdin)
//
// The Manager.Rotate method is called automatically by NewWriter before
// creating a new snapshot file, ensuring stale files are pruned on each
// capture session.
package rotate
