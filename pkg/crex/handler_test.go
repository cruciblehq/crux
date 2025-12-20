package crex

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewHandler(t *testing.T) {
	h := NewHandler()

	if h == nil {
		t.Fatal("NewHandler() returned nil")
	}

	// Verify default level is Info
	if h.Level() != slog.LevelInfo {
		t.Errorf("Level() = %v, want %v", h.Level(), slog.LevelInfo)
	}

	// Verify default stream is os.Stderr (can't test exact value but can verify it's not nil)
	if h.Stream() == nil {
		t.Error("Stream() is nil, want non-nil default")
	}
}

func TestNewHandlerWithLevel(t *testing.T) {
	level := slog.LevelWarn
	h := NewHandlerWithLevel(level)

	if h == nil {
		t.Fatal("NewHandlerWithLevel() returned nil")
	}

	if h.Level() != level {
		t.Errorf("Level() = %v, want %v", h.Level(), level)
	}
}

func TestHandler_SetLevel(t *testing.T) {
	h := NewHandler()
	newLevel := slog.LevelDebug

	result := h.SetLevel(newLevel)

	if result != h {
		t.Error("SetLevel() did not return the handler for chaining")
	}

	if h.Level() != newLevel {
		t.Errorf("Level() = %v, want %v", h.Level(), newLevel)
	}
}

func TestHandler_SetStream(t *testing.T) {
	h := NewHandler()
	var buf bytes.Buffer

	result := h.SetStream(&buf)

	if result != h {
		t.Error("SetStream() did not return the handler for chaining")
	}

	if h.Stream() != &buf {
		t.Error("Stream() is not the set buffer")
	}
}

func TestHandler_SetFormatter(t *testing.T) {
	h := NewHandler()
	formatter := NewPrettyFormatter(false)

	result := h.SetFormatter(formatter)

	if result != h {
		t.Error("SetFormatter() did not return the handler for chaining")
	}

	if h.Formatter() != formatter {
		t.Error("Formatter() is not the set formatter")
	}
}

func TestHandler_Enabled(t *testing.T) {
	tests := []struct {
		name         string
		handlerLevel slog.Level
		recordLevel  slog.Level
		want         bool
	}{
		{"debug handler, debug record", slog.LevelDebug, slog.LevelDebug, true},
		{"debug handler, info record", slog.LevelDebug, slog.LevelInfo, true},
		{"info handler, debug record", slog.LevelInfo, slog.LevelDebug, false},
		{"info handler, info record", slog.LevelInfo, slog.LevelInfo, true},
		{"warn handler, info record", slog.LevelWarn, slog.LevelInfo, false},
		{"warn handler, error record", slog.LevelWarn, slog.LevelError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandlerWithLevel(tt.handlerLevel)

			got := h.Enabled(context.Background(), tt.recordLevel)
			if got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_Handle(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler().
		SetStream(&buf).
		SetFormatter(NewPrettyFormatter(false))

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	err := h.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	// With formatter set, should write immediately
	if buf.Len() == 0 {
		t.Error("Handle() did not write to buffer")
	}

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("output does not contain message: %s", output)
	}
}

func TestHandler_HandleBuffering(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler().SetStream(&buf)
	// No formatter set - should buffer

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	err := h.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	// Without formatter, should buffer (no output yet)
	if buf.Len() > 0 {
		t.Error("Handle() wrote to buffer before formatter was set")
	}

	// Now set formatter and flush
	h.SetFormatter(NewPrettyFormatter(false))
	flushed, err := h.Flush()
	if err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	if !flushed {
		t.Error("Flush() returned false, want true")
	}

	if buf.Len() == 0 {
		t.Error("Flush() did not write buffered records")
	}

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("output does not contain message: %s", output)
	}
}

func TestHandler_FlushMultiple(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler().SetStream(&buf)
	// No formatter - buffer records

	// Handle multiple records
	for i := 0; i < 3; i++ {
		record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
		if err := h.Handle(context.Background(), record); err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
	}

	// Set formatter and flush
	h.SetFormatter(NewPrettyFormatter(false))
	flushed, err := h.Flush()
	if err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	if !flushed {
		t.Error("Flush() returned false, want true")
	}

	output := buf.String()
	count := strings.Count(output, "test")
	if count != 3 {
		t.Errorf("output contains %d occurrences of 'test', want 3", count)
	}
}

func TestHandler_FlushEmpty(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler().
		SetStream(&buf).
		SetFormatter(NewPrettyFormatter(false))

	// Flush without handling any records
	flushed, err := h.Flush()
	if err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	if !flushed {
		t.Error("Flush() returned false, want true (formatter is set)")
	}

	if buf.Len() > 0 {
		t.Error("Flush() wrote to buffer with no records")
	}
}

func TestHandler_FlushNoFormatter(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler().SetStream(&buf)
	// No formatter set

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	if err := h.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	// Flush without formatter should return false
	flushed, err := h.Flush()
	if err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	if flushed {
		t.Error("Flush() returned true without formatter, want false")
	}

	if buf.Len() > 0 {
		t.Error("Flush() wrote to buffer without formatter")
	}
}

func TestHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewPrettyFormatter(false)
	formatter.SetVerbose(true)

	h := NewHandler().
		SetStream(&buf).
		SetFormatter(formatter)

	attrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	}

	h2 := h.WithAttrs(attrs)

	// Should return a new handler
	if h2 == h {
		t.Error("WithAttrs() returned the same handler instance")
	}

	// Test that attributes are included in output
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	if err := h2.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "key1") {
		t.Error("output does not contain key1 attribute")
	}
}

func TestHandler_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewPrettyFormatter(false)
	formatter.SetVerbose(true)

	h := NewHandler().
		SetStream(&buf).
		SetFormatter(formatter)

	h2 := h.WithGroup("group1")

	// Should return a new handler
	if h2 == h {
		t.Error("WithGroup() returned the same handler instance")
	}

	// Verify group is preserved in output
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	if err := h2.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "group1") {
		t.Error("output does not contain group1")
	}
}

func TestHandler_WithGroup_Empty(t *testing.T) {
	h := NewHandler()

	// Empty group names should be ignored
	h2 := h.WithGroup("")

	// Should return a handler (possibly the same or a new one)
	if h2 == nil {
		t.Error("WithGroup(\"\") returned nil")
	}
}

func TestHandler_WithGroup_Nested(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewPrettyFormatter(false)
	formatter.SetVerbose(true)

	h := NewHandler().
		SetStream(&buf).
		SetFormatter(formatter)

	h2 := h.WithGroup("group1")
	h3 := h2.WithGroup("group2")

	// Verify nested groups in output
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	if err := h3.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "group1") {
		t.Error("output does not contain group1")
	}
	if !strings.Contains(output, "group2") {
		t.Error("output does not contain group2")
	}
}

func TestHandler_ConcurrentHandle(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler().
		SetStream(&buf).
		SetFormatter(NewPrettyFormatter(false))

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
			record.AddAttrs(slog.Int("n", n))
			if err := h.Handle(context.Background(), record); err != nil {
				t.Errorf("Handle() error = %v", err)
			}
		}(i)
	}
	wg.Wait()

	// Verify all records were handled
	output := buf.String()
	count := strings.Count(output, "test")
	if count != 10 {
		t.Errorf("output contains %d records, want 10", count)
	}
}

func TestHandler_ConcurrentFlush(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler().SetStream(&buf)

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	if err := h.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	// Set formatter
	h.SetFormatter(NewPrettyFormatter(false))

	// Multiple concurrent flushes should be safe
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.Flush()
		}()
	}
	wg.Wait()

	// Should have output from at least one flush
	if buf.Len() == 0 {
		t.Error("no output after concurrent flushes")
	}
}

func TestHandler_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandlerWithLevel(slog.LevelWarn).
		SetStream(&buf).
		SetFormatter(NewPrettyFormatter(false))

	// Debug and Info should not be enabled
	record1 := slog.NewRecord(time.Now(), slog.LevelDebug, "debug", 0)
	if h.Enabled(context.Background(), record1.Level) {
		t.Error("Debug level is enabled when level is Warn")
	}
	// Don't call Handle for disabled levels

	record2 := slog.NewRecord(time.Now(), slog.LevelInfo, "info", 0)
	if h.Enabled(context.Background(), record2.Level) {
		t.Error("Info level is enabled when level is Warn")
	}
	// Don't call Handle for disabled levels

	// Warn should be enabled and written
	record3 := slog.NewRecord(time.Now(), slog.LevelWarn, "warn", 0)
	if !h.Enabled(context.Background(), record3.Level) {
		t.Error("Warn level is not enabled when level is Warn")
	}
	if err := h.Handle(context.Background(), record3); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "debug") {
		t.Error("output contains debug message (should not be present)")
	}
	if strings.Contains(output, "info") {
		t.Error("output contains info message (should not be present)")
	}
	if !strings.Contains(output, "warn") {
		t.Error("output does not contain warn message")
	}
}
