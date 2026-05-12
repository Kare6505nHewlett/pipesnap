package cli

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// Config holds the parsed CLI configuration for pipesnap.
type Config struct {
	Mode       string
	OutputFile string
	InputFile  string
	ChunkDelay time.Duration
	Verbose    bool
}

// ParseFlags parses command-line arguments and returns a Config.
// It exits with usage information on invalid input.
func ParseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("pipesnap", flag.ContinueOnError)

	output := fs.String("o", "", "output snapshot file (used in snap mode)")
	input := fs.String("i", "", "input snapshot file (used in replay mode)")
	delay := fs.Duration("delay", 0, "delay between replayed chunks (e.g. 10ms)")
	verbose := fs.Bool("v", false, "enable verbose logging")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pipesnap <snap|replay> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Modes:\n")
		fmt.Fprintf(os.Stderr, "  snap    Read stdin and write a snapshot file\n")
		fmt.Fprintf(os.Stderr, "  replay  Read a snapshot file and write to stdout\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	if len(args) < 1 {
		fs.Usage()
		return nil, fmt.Errorf("mode required: snap or replay")
	}

	mode := args[0]
	if mode != "snap" && mode != "replay" {
		return nil, fmt.Errorf("unknown mode %q: expected snap or replay", mode)
	}

	if err := fs.Parse(args[1:]); err != nil {
		return nil, err
	}

	cfg := &Config{
		Mode:       mode,
		OutputFile: *output,
		InputFile:  *input,
		ChunkDelay: *delay,
		Verbose:    *verbose,
	}

	if mode == "snap" && cfg.OutputFile == "" {
		return nil, fmt.Errorf("snap mode requires -o <output file>")
	}
	if mode == "replay" && cfg.InputFile == "" {
		return nil, fmt.Errorf("replay mode requires -i <input file>")
	}

	return cfg, nil
}
