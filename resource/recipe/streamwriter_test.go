package recipe

import (
	"context"
	"log/slog"
	"testing"
)

// Collects log records emitted during a test.
type captureHandler struct {
	records *[]slog.Record
}

func (h captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h captureHandler) Handle(_ context.Context, r slog.Record) error {
	*h.records = append(*h.records, r)
	return nil
}

func (h captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h captureHandler) WithGroup(string) slog.Handler { return h }

func TestStreamWriter(t *testing.T) {
	var records []slog.Record
	prev := slog.Default()
	slog.SetDefault(slog.New(captureHandler{records: &records}))
	defer slog.SetDefault(prev)

	w := newStreamWriter(context.Background(), slog.LevelInfo, "stdout")

	// Partial write without a newline must not emit yet.
	if n, err := w.Write([]byte("hello ")); err != nil || n != 6 {
		t.Fatalf("Write partial = (%d, %v)", n, err)
	}
	if len(records) != 0 {
		t.Fatalf("partial line emitted %d records", len(records))
	}

	// Completing the line plus a second full line emits two records.
	if _, err := w.Write([]byte("world\nsecond\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("want 2 records, got %d", len(records))
	}
	if records[0].Message != "hello world" || records[1].Message != "second" {
		t.Fatalf("unexpected messages: %q, %q", records[0].Message, records[1].Message)
	}

	var stream string
	records[0].Attrs(func(a slog.Attr) bool {
		if a.Key == "stream" {
			stream = a.Value.String()
		}
		return true
	})
	if stream != "stdout" {
		t.Fatalf("stream attr = %q, want stdout", stream)
	}
}
