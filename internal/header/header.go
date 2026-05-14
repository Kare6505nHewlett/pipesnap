// Package header provides utilities for reading and writing a metadata
// header at the start of a snapshot file. The header stores key/value
// pairs (e.g. capture timestamp, hostname, pipesnap version) encoded as
// a single JSON line followed by a newline delimiter.
package header

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

const delimiter = '\n'

// Meta holds the metadata written at the top of every snapshot file.
type Meta struct {
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Hostname  string    `json:"hostname,omitempty"`
	Extra     map[string]string `json:"extra,omitempty"`
}

// Write serialises m as a single JSON line into w.
// It must be called before any snapshot chunk data is written.
func Write(w io.Writer, m Meta) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("header: marshal: %w", err)
	}
	data = append(data, delimiter)
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("header: write: %w", err)
	}
	return nil
}

// Read deserialises the first JSON line from r into a Meta value.
// After Read returns the reader is positioned immediately after the
// header line, ready for snapshot chunk data.
func Read(r io.Reader) (Meta, error) {
	var buf []byte
	tmp := make([]byte, 1)
	for {
		n, err := r.Read(tmp)
		if n > 0 {
			if tmp[0] == delimiter {
				break
			}
			buf = append(buf, tmp[0])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return Meta{}, fmt.Errorf("header: read: %w", err)
		}
	}
	if len(buf) == 0 {
		return Meta{}, fmt.Errorf("header: empty header")
	}
	var m Meta
	if err := json.Unmarshal(buf, &m); err != nil {
		return Meta{}, fmt.Errorf("header: unmarshal: %w", err)
	}
	return m, nil
}
