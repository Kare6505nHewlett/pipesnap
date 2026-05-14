// Package checkpoint provides functionality to save and restore
// stream positions, enabling resumable snapshot sessions.
package checkpoint

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

// ErrNoCheckpoint is returned when no checkpoint file exists.
var ErrNoCheckpoint = errors.New("checkpoint: no checkpoint file found")

// State holds the persisted state of a snapshot session.
type State struct {
	SnapshotFile string    `json:"snapshot_file"`
	BytesWritten int64     `json:"bytes_written"`
	ChunksWritten int64    `json:"chunks_written"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Save writes the checkpoint state to the given file path.
func Save(path string, s State) error {
	s.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Load reads and parses a checkpoint state from the given file path.
// Returns ErrNoCheckpoint if the file does not exist.
func Load(path string) (State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, ErrNoCheckpoint
		}
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, err
	}
	return s, nil
}

// Remove deletes the checkpoint file at the given path.
// Returns nil if the file does not exist.
func Remove(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
