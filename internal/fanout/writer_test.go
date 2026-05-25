package fanout_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/yourorg/pipesnap/internal/fanout"
)

func makeFanout(t *testing.T, dests ...fanout.Destination) *fanout.Fanout {
	t.Helper()
	f, err := fanout.New(dests)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestNewWriterRejectsNil(t *testing.T) {
	_, err := fanout.NewWriter(nil)
	if err == nil {
		t.Fatal("expected error for nil fanout")
	}
}

func TestWriterPassesData(t *testing.T) {
	var buf bytes.Buffer
	f := makeFanout(t, fanout.Destination{Name: "x", Dst: &buf})
	w, err := fanout.NewWriter(f)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "ping" {
		t.Errorf("got %q", buf.String())
	}
}

func TestWriterCloseCallsClosers(t *testing.T) {
	c := &trackCloser{}
	var plain bytes.Buffer
	f := makeFanout(t,
		fanout.Destination{Name: "closeable", Dst: c},
		fanout.Destination{Name: "plain", Dst: &plain},
	)
	w, _ := fanout.NewWriter(f)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if !c.closed {
		t.Error("expected closer to be called")
	}
}

func TestWriterCloseReturnsError(t *testing.T) {
	c := &trackCloser{err: errors.New("close failed")}
	f := makeFanout(t, fanout.Destination{Name: "bad", Dst: c})
	w, _ := fanout.NewWriter(f)
	if err := w.Close(); err == nil {
		t.Fatal("expected close error")
	}
}

// trackCloser is an io.WriteCloser that records whether Close was called.
type trackCloser struct {
	closed bool
	err    error
}

func (t *trackCloser) Write(p []byte) (int, error) { return len(p), nil }
func (t *trackCloser) Close() error {
	t.closed = true
	return t.err
}
