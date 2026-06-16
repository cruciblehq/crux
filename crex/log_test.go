package crex

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

// Builds a logger backed by a pretty formatter writing to a buffer, returning
// the logger, the handler for flushing, and the buffer for assertions.
func newBufferLogger(verbose bool) (*slog.Logger, Handler, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	f := NewPrettyFormatter(false)
	f.Verbose = verbose
	h := NewHandler()
	h.SetStream(buf)
	h.SetFormatter(f)
	return slog.New(h), h, buf
}

func TestLogError_ClassfulError(t *testing.T) {
	logger, h, buf := newBufferLogger(true)

	err := UserError("invalid type", "widget").
		Recovery("Use a valid type.").
		Err()

	LogError(logger, err)
	if _, ferr := h.Flush(); ferr != nil {
		t.Fatalf("Flush() error = %v", ferr)
	}

	got := buf.String()
	for _, want := range []string{"invalid type", "widget", "Use a valid type.", "class=user"} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q missing %q", got, want)
		}
	}
}

func TestLogError_ClasslessWrapped(t *testing.T) {
	logger, h, buf := newBufferLogger(true)

	sentinel := New("read failed")
	err := Wrapf(sentinel, errors.New("disk gone"), "cannot read manifest")

	LogError(logger, err)
	if _, ferr := h.Flush(); ferr != nil {
		t.Fatalf("Flush() error = %v", ferr)
	}

	got := buf.String()
	// Routed through the structured formatter: headline is the sentinel and the
	// reason is appended, all on a single line.
	for _, want := range []string{"read failed", "cannot read manifest", "disk gone"} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q missing %q", got, want)
		}
	}
	if strings.Count(strings.TrimRight(got, "\n"), "\n") != 0 {
		t.Errorf("expected single-line output, got %q", got)
	}
}

func TestLogError_ForeignError(t *testing.T) {
	logger, h, buf := newBufferLogger(true)

	LogError(logger, errors.New("boom"))
	if _, ferr := h.Flush(); ferr != nil {
		t.Fatalf("Flush() error = %v", ferr)
	}

	got := buf.String()
	// A foreign error is adapted to the structured form under the unknown class.
	for _, want := range []string{"boom", "class=unknown"} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q missing %q", got, want)
		}
	}
}

func TestLogError_Nil(t *testing.T) {
	logger, h, buf := newBufferLogger(false)

	LogError(logger, nil)
	if _, ferr := h.Flush(); ferr != nil {
		t.Fatalf("Flush() error = %v", ferr)
	}

	if got := buf.String(); got != "" {
		t.Errorf("expected no output for nil error, got %q", got)
	}
}
