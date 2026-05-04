package crex

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestNewPrettyFormatter(t *testing.T) {
	f := NewPrettyFormatter(true)

	if !f.UseColor {
		t.Error("UseColor = false, want true")
	}
	if f.Verbose() {
		t.Error("Verbose() = true, want false (default)")
	}
}

func TestPrettyFormatter_SetVerbose(t *testing.T) {
	f := NewPrettyFormatter(false)
	f.SetVerbose(true)

	if !f.Verbose() {
		t.Error("Verbose() = false after SetVerbose(true)")
	}
}

func TestPrettyFormatter_Write_Simple(t *testing.T) {
	f := NewPrettyFormatter(false)
	var buf bytes.Buffer

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	rctx := &RecordContext{Record: record}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "[info]") {
		t.Errorf("output missing level: %q", got)
	}
	if !strings.Contains(got, "test message") {
		t.Errorf("output missing message: %q", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Error("output does not end with newline")
	}
}

func TestPrettyFormatter_Write_WithColor(t *testing.T) {
	f := NewPrettyFormatter(true)
	var buf bytes.Buffer

	record := slog.NewRecord(time.Now(), slog.LevelError, "error occurred", 0)
	rctx := &RecordContext{Record: record}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got := buf.String()
	// Should contain ANSI color codes
	if !strings.Contains(got, "\033[") {
		t.Error("output missing ANSI color codes")
	}
	if !strings.Contains(got, colorReset) {
		t.Error("output missing color reset")
	}
}

func TestPrettyFormatter_Write_WithGroups(t *testing.T) {
	f := NewPrettyFormatter(false)
	var buf bytes.Buffer

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	rctx := &RecordContext{
		Record: record,
		Groups: []string{"group1", "group2"},
	}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got := buf.String()
	if !strings.HasPrefix(got, "group1.group2") {
		t.Errorf("output missing groups: %q", got)
	}
}

func TestPrettyFormatter_Write_WithAttributes_NonVerbose(t *testing.T) {
	f := NewPrettyFormatter(false)
	f.SetVerbose(false)
	var buf bytes.Buffer

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	record.AddAttrs(slog.String("key", "value"))
	rctx := &RecordContext{Record: record}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got := buf.String()
	// Non-verbose should not show regular attributes
	if strings.Contains(got, "key=value") {
		t.Errorf("non-verbose output contains attribute: %q", got)
	}
}

func TestPrettyFormatter_Write_WithAttributes_Verbose(t *testing.T) {
	f := NewPrettyFormatter(false)
	f.SetVerbose(true)
	var buf bytes.Buffer

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	record.AddAttrs(slog.String("key", "value"))
	rctx := &RecordContext{Record: record}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got := buf.String()
	// Verbose should show attributes
	if !strings.Contains(got, "key=value") {
		t.Errorf("verbose output missing attribute: %q", got)
	}
}

func TestPrettyFormatter_Write_CrexError(t *testing.T) {
	f := NewPrettyFormatter(false)
	var buf bytes.Buffer

	// Create a crex error and add its LogValue directly
	crexErr := UserError("operation failed", "invalid input").Err().(*Error)
	record := slog.NewRecord(time.Now(), slog.LevelError, "task failed", 0)
	record.AddAttrs(slog.Any("error", crexErr))
	rctx := &RecordContext{Record: record}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "task failed") {
		t.Errorf("output missing message: %q", got)
	}
	if !strings.Contains(got, ": invalid input") {
		t.Errorf("output missing reason: %q", got)
	}
}

func TestPrettyFormatter_Write_CrexError_WithFallback(t *testing.T) {
	f := NewPrettyFormatter(false)
	var buf bytes.Buffer

	crexErr := UserError("operation failed", "invalid type").
		Fallback("Use widget or service").
		Err().(*Error)

	record := slog.NewRecord(time.Now(), slog.LevelError, "build failed", 0)
	record.AddAttrs(slog.Any("error", crexErr))
	rctx := &RecordContext{Record: record}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, ": invalid type") {
		t.Errorf("output missing reason: %q", got)
	}
	if !strings.Contains(got, ". Use widget or service") {
		t.Errorf("output missing fallback: %q", got)
	}
}

func TestPrettyFormatter_Write_CrexError_Verbose(t *testing.T) {
	f := NewPrettyFormatter(false).SetVerbose(true)
	var buf bytes.Buffer

	crexErr := SystemError("network error", "connection timeout").
		Detail("host", "example.com").
		Err().(*Error)

	record := slog.NewRecord(time.Now(), slog.LevelError, "request failed", 0)
	record.AddAttrs(slog.Any("error", crexErr))
	rctx := &RecordContext{Record: record}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got := buf.String()
	// Verbose should include error details
	if !strings.Contains(got, "class=system") {
		t.Errorf("verbose output missing class: %q", got)
	}
	if !strings.Contains(got, "host=example.com") {
		t.Errorf("verbose output missing details: %q", got)
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		value slog.Value
		want  string
	}{
		{
			name:  "string",
			value: slog.StringValue("test"),
			want:  "test",
		},
		{
			name:  "int",
			value: slog.IntValue(42),
			want:  "42",
		},
		{
			name:  "duration",
			value: slog.DurationValue(5 * time.Second),
			want:  "5s",
		},
		{
			name:  "time",
			value: slog.TimeValue(time.Date(2024, 1, 1, 15, 4, 5, 0, time.UTC)),
			want:  "15:04:05",
		},
		{
			name: "group",
			value: slog.GroupValue(
				slog.String("a", "1"),
				slog.String("b", "2"),
			),
			want: "{a=1 b=2}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.value)
			if got != tt.want {
				t.Errorf("formatValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAsCrexError(t *testing.T) {
	t.Run("valid crex error", func(t *testing.T) {
		crexErr := UserError("test", "reason").Err()
		val := crexErr.(*Error).LogValue()

		errMap, ok := asCrexError(val)
		if !ok {
			t.Error("asCrexError() = false, want true")
		}
		if errMap == nil {
			t.Fatal("asCrexError() returned nil map")
		}

		if _, hasClass := errMap["class"]; !hasClass {
			t.Error("error map missing 'class' key")
		}
		if _, hasDesc := errMap["description"]; !hasDesc {
			t.Error("error map missing 'description' key")
		}
	})

	t.Run("non-group value", func(t *testing.T) {
		val := slog.StringValue("test")
		_, ok := asCrexError(val)
		if ok {
			t.Error("asCrexError() = true for non-group, want false")
		}
	})

	t.Run("group without required keys", func(t *testing.T) {
		val := slog.GroupValue(
			slog.String("other", "value"),
		)
		_, ok := asCrexError(val)
		if ok {
			t.Error("asCrexError() = true for incomplete group, want false")
		}
	})
}

func TestLevelColor(t *testing.T) {
	tests := []struct {
		level slog.Level
		want  string
	}{
		{slog.LevelDebug, "\033[37m"},
		{slog.LevelInfo, "\033[32m"},
		{slog.LevelWarn, "\033[33m"},
		{slog.LevelError, "\033[31m"},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			got := levelColor(tt.level)
			if got != tt.want {
				t.Errorf("levelColor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrettyFormatter_LogLevels(t *testing.T) {
	levels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	f := NewPrettyFormatter(false)

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			var buf bytes.Buffer
			record := slog.NewRecord(time.Now(), level, "test", 0)
			rctx := &RecordContext{Record: record}

			err := f.Write(&buf, rctx)
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			got := buf.String()
			levelStr := "[" + strings.ToLower(level.String()) + "]"
			if !strings.Contains(got, levelStr) {
				t.Errorf("output missing level string %q: %q", levelStr, got)
			}
		})
	}
}
