// Package audit records a structured log entry for every chunk that passes
// through the pipeline. Each entry captures the chunk size, a truncated
// preview, and an optional label so operators can trace data flow.
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

const defaultPreviewLen = 64

// Entry is the structured record written for each observed chunk.
type Entry struct {
	At        time.Time `json:"at"`
	Label     string    `json:"label,omitempty"`
	Bytes     int       `json:"bytes"`
	Preview   string    `json:"preview"`
	Truncated bool      `json:"truncated,omitempty"`
}

// Logger writes audit entries to an io.Writer as newline-delimited JSON.
type Logger struct {
	dst        io.Writer
	label      string
	previewLen int
}

// New creates a Logger that writes to dst.
// label is attached to every entry; previewLen controls how many bytes of
// the raw chunk body are captured (0 uses the default of 64).
func New(dst io.Writer, label string, previewLen int) (*Logger, error) {
	if dst == nil {
		return nil, fmt.Errorf("audit: dst must not be nil")
	}
	if previewLen <= 0 {
		previewLen = defaultPreviewLen
	}
	return &Logger{dst: dst, label: label, previewLen: previewLen}, nil
}

// Record writes a single audit entry for data.
func (l *Logger) Record(data []byte) error {
	preview := data
	truncated := false
	if len(preview) > l.previewLen {
		preview = preview[:l.previewLen]
		truncated = true
	}
	entry := Entry{
		At:        time.Now().UTC(),
		Label:     l.label,
		Bytes:     len(data),
		Preview:   string(preview),
		Truncated: truncated,
	}
	b, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: marshal: %w", err)
	}
	b = append(b, '\n')
	_, err = l.dst.Write(b)
	return err
}
