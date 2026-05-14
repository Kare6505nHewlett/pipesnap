package header_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"pipesnap/internal/header"
)

func TestWriteAndRead(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	m := header.Meta{
		Version:   "1.0.0",
		CreatedAt: now,
		Hostname:  "testhost",
		Extra:     map[string]string{"env": "ci"},
	}

	var buf bytes.Buffer
	if err := header.Write(&buf, m); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := header.Read(&buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if got.Version != m.Version {
		t.Errorf("Version: got %q, want %q", got.Version, m.Version)
	}
	if !got.CreatedAt.Equal(m.CreatedAt) {
		t.Errorf("CreatedAt: got %v, want %v", got.CreatedAt, m.CreatedAt)
	}
	if got.Hostname != m.Hostname {
		t.Errorf("Hostname: got %q, want %q", got.Hostname, m.Hostname)
	}
	if got.Extra["env"] != "ci" {
		t.Errorf("Extra[env]: got %q, want %q", got.Extra["env"], "ci")
	}
}

func TestReadLeavesBodyIntact(t *testing.T) {
	body := "chunk data follows"
	var buf bytes.Buffer
	header.Write(&buf, header.Meta{Version: "0.1", CreatedAt: time.Now()})
	buf.WriteString(body)

	if _, err := header.Read(&buf); err != nil {
		t.Fatalf("Read: %v", err)
	}
	remaining := buf.String()
	if remaining != body {
		t.Errorf("remaining body: got %q, want %q", remaining, body)
	}
}

func TestReadEmptyReturnsError(t *testing.T) {
	_, err := header.Read(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty reader, got nil")
	}
}

func TestReadInvalidJSONReturnsError(t *testing.T) {
	_, err := header.Read(strings.NewReader("{not valid json}\n"))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
