package replay

import (
	"io"
	"time"

	"github.com/user/pipesnap/internal/snapshot"
)

// Options controls replay behavior.
type Options struct {
	// Delay between chunks. Zero means no delay.
	ChunkDelay time.Duration
	// Writer to replay chunks into. Defaults to os.Stdout if nil.
	Dest io.Writer
}

// Replayer reads a snapshot file and writes chunks to a destination writer.
type Replayer struct {
	reader *snapshot.Reader
	opts   Options
}

// New creates a Replayer that reads from the snapshot at path.
func New(path string, opts Options) (*Replayer, error) {
	r, err := snapshot.OpenReader(path)
	if err != nil {
		return nil, err
	}
	return &Replayer{reader: r, opts: opts}, nil
}

// Run replays all chunks to the destination writer.
// It returns the total number of bytes written and any error encountered.
func (r *Replayer) Run() (int64, error) {
	var total int64
	first := true

	for {
		chunk, err := r.reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return total, err
		}

		if !first && r.opts.ChunkDelay > 0 {
			time.Sleep(r.opts.ChunkDelay)
		}
		first = false

		n, err := r.opts.Dest.Write(chunk)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

// Close releases resources held by the underlying reader.
func (r *Replayer) Close() error {
	return r.reader.Close()
}
