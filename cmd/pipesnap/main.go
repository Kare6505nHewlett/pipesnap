package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"pipesnap/internal/cli"
	"pipesnap/internal/replay"
	"pipesnap/internal/snapshot"
)

func main() {
	cfg, err := cli.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	logger := log.New(io.Discard, "", 0)
	if cfg.Verbose {
		logger = log.New(os.Stderr, "[pipesnap] ", log.LstdFlags)
	}

	switch cfg.Mode {
	case "snap":
		if err := runSnap(cfg, logger); err != nil {
			fmt.Fprintf(os.Stderr, "snap error: %v\n", err)
			os.Exit(1)
		}
	case "replay":
		if err := runReplay(cfg, logger); err != nil {
			fmt.Fprintf(os.Stderr, "replay error: %v\n", err)
			os.Exit(1)
		}
	}
}

func runSnap(cfg *cli.Config, logger *log.Logger) error {
	logger.Printf("snapping stdin to %s", cfg.OutputFile)

	w, err := snapshot.NewWriter(cfg.OutputFile)
	if err != nil {
		return fmt.Errorf("open snapshot writer: %w", err)
	}
	defer w.Close()

	n, err := io.Copy(w, os.Stdin)
	if err != nil {
		return fmt.Errorf("writing snapshot: %w", err)
	}

	logger.Printf("snapshot complete: %d bytes written", n)
	return nil
}

func runReplay(cfg *cli.Config, logger *log.Logger) error {
	logger.Printf("replaying %s to stdout", cfg.InputFile)

	r, err := replay.New(cfg.InputFile, cfg.ChunkDelay)
	if err != nil {
		return fmt.Errorf("open replay: %w", err)
	}
	defer r.Close()

	n, err := io.Copy(os.Stdout, r)
	if err != nil {
		return fmt.Errorf("replaying snapshot: %w", err)
	}

	logger.Printf("replay complete: %d bytes written", n)
	return nil
}
