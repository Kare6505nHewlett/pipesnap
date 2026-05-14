// Package rotate provides snapshot file rotation based on size or count limits.
package rotate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Config holds rotation policy settings.
type Config struct {
	// MaxFiles is the maximum number of snapshot files to keep (0 = unlimited).
	MaxFiles int
	// MaxBytes is the maximum total size in bytes before rotating (0 = unlimited).
	MaxBytes int64
	// Dir is the directory where snapshots are stored.
	Dir string
	// Prefix is the filename prefix for snapshot files.
	Prefix string
}

// Manager manages snapshot file rotation.
type Manager struct {
	cfg Config
}

// New creates a new rotation Manager with the given config.
func New(cfg Config) *Manager {
	return &Manager{cfg: cfg}
}

// NextPath returns a new timestamped snapshot file path.
func (m *Manager) NextPath() string {
	ts := time.Now().UTC().Format("20060102T150405Z")
	name := fmt.Sprintf("%s%s.snap", m.cfg.Prefix, ts)
	return filepath.Join(m.cfg.Dir, name)
}

// Rotate removes old snapshot files according to the configured policy.
// It returns the list of removed file paths.
func (m *Manager) Rotate() ([]string, error) {
	files, err := m.listSnapshots()
	if err != nil {
		return nil, err
	}

	var removed []string

	if m.cfg.MaxFiles > 0 {
		for len(files) > m.cfg.MaxFiles {
			target := files[0]
			if err := os.Remove(target.path); err != nil && !os.IsNotExist(err) {
				return removed, fmt.Errorf("rotate: remove %s: %w", target.path, err)
			}
			removed = append(removed, target.path)
			files = files[1:]
		}
	}

	if m.cfg.MaxBytes > 0 {
		var total int64
		for _, f := range files {
			total += f.size
		}
		for total > m.cfg.MaxBytes && len(files) > 0 {
			target := files[0]
			if err := os.Remove(target.path); err != nil && !os.IsNotExist(err) {
				return removed, fmt.Errorf("rotate: remove %s: %w", target.path, err)
			}
			removed = append(removed, target.path)
			total -= target.size
			files = files[1:]
		}
	}

	return removed, nil
}

type fileEntry struct {
	path    string
	size    int64
	modTime time.Time
}

func (m *Manager) listSnapshots() ([]fileEntry, error) {
	entries, err := os.ReadDir(m.cfg.Dir)
	if err != nil {
		return nil, fmt.Errorf("rotate: read dir %s: %w", m.cfg.Dir, err)
	}

	var files []fileEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, m.cfg.Prefix) || !strings.HasSuffix(name, ".snap") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileEntry{
			path:    filepath.Join(m.cfg.Dir, name),
			size:    info.Size(),
			modTime: info.ModTime(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})
	return files, nil
}
