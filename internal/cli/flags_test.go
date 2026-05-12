package cli

import (
	"testing"
	"time"
)

func TestParseFlagsSnapMode(t *testing.T) {
	cfg, err := ParseFlags([]string{"snap", "-o", "out.snap", "-v"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != "snap" {
		t.Errorf("expected mode snap, got %q", cfg.Mode)
	}
	if cfg.OutputFile != "out.snap" {
		t.Errorf("expected output file out.snap, got %q", cfg.OutputFile)
	}
	if !cfg.Verbose {
		t.Error("expected verbose to be true")
	}
}

func TestParseFlagsReplayMode(t *testing.T) {
	cfg, err := ParseFlags([]string{"replay", "-i", "in.snap", "-delay", "20ms"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != "replay" {
		t.Errorf("expected mode replay, got %q", cfg.Mode)
	}
	if cfg.InputFile != "in.snap" {
		t.Errorf("expected input file in.snap, got %q", cfg.InputFile)
	}
	if cfg.ChunkDelay != 20*time.Millisecond {
		t.Errorf("expected delay 20ms, got %v", cfg.ChunkDelay)
	}
}

func TestParseFlagsMissingMode(t *testing.T) {
	_, err := ParseFlags([]string{})
	if err == nil {
		t.Fatal("expected error for missing mode")
	}
}

func TestParseFlagsUnknownMode(t *testing.T) {
	_, err := ParseFlags([]string{"stream", "-o", "out.snap"})
	if err == nil {
		t.Fatal("expected error for unknown mode")
	}
}

func TestParseFlagsSnapMissingOutput(t *testing.T) {
	_, err := ParseFlags([]string{"snap"})
	if err == nil {
		t.Fatal("expected error when -o is missing in snap mode")
	}
}

func TestParseFlagsReplayMissingInput(t *testing.T) {
	_, err := ParseFlags([]string{"replay"})
	if err == nil {
		t.Fatal("expected error when -i is missing in replay mode")
	}
}
