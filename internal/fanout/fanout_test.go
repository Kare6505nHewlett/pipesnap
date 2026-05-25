package fanout_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/yourorg/pipesnap/internal/fanout"
)

func TestNewRejectsNoDests(t *testing.T) {
	_, err := fanout.New(nil)
	if err == nil {
		t.Fatal("expected error for empty destinations")
	}
}

func TestNewRejectsNilWriter(t *testing.T) {
	_, err := fanout.New([]fanout.Destination{
		{Name: "a", Dst: nil},
	})
	if err == nil {
		t.Fatal("expected error for nil writer")
	}
}

func TestWriteFansOutToAll(t *testing.T) {
	var a, b bytes.Buffer
	f, err := fanout.New([]fanout.Destination{
		{Name: "a", Dst: &a},
		{Name: "b", Dst: &b},
	})
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("hello")
	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}
	if a.String() != "hello" {
		t.Errorf("a: got %q", a.String())
	}
	if b.String() != "hello" {
		t.Errorf("b: got %q", b.String())
	}
}

func TestFilterDropsChunkForOneDest(t *testing.T) {
	var a, b bytes.Buffer
	f, err := fanout.New([]fanout.Destination{
		{Name: "a", Dst: &a, Filter: func(p []byte) bool {
			return strings.Contains(string(p), "keep")
		}},
		{Name: "b", Dst: &b},
	})
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("drop this"))
	f.Write([]byte("keep this"))

	if a.String() != "keep this" {
		t.Errorf("a: got %q, want %q", a.String(), "keep this")
	}
	if b.String() != "drop thiskeep this" {
		t.Errorf("b: got %q", b.String())
	}
}

func TestWriteReturnsFirstError(t *testing.T) {
	boom := &errWriter{err: errors.New("boom")}
	var good bytes.Buffer
	f, err := fanout.New([]fanout.Destination{
		{Name: "bad", Dst: boom},
		{Name: "good", Dst: &good},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("data"))
	if err == nil {
		t.Fatal("expected error")
	}
	// good destination should still have received the data
	if good.String() != "data" {
		t.Errorf("good: got %q", good.String())
	}
}

func TestLen(t *testing.T) {
	var a, b, c bytes.Buffer
	f, _ := fanout.New([]fanout.Destination{
		{Name: "a", Dst: &a},
		{Name: "b", Dst: &b},
		{Name: "c", Dst: &c},
	})
	if f.Len() != 3 {
		t.Errorf("Len() = %d, want 3", f.Len())
	}
}

type errWriter struct{ err error }

func (e *errWriter) Write(p []byte) (int, error) { return 0, e.err }
