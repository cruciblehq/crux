package crex

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestNewJSONFormatter(t *testing.T) {
	f := NewJSONFormatter()
	if f == nil {
		t.Fatal("NewJSONFormatter() returned nil")
	}
}

func TestJSONFormatter_Write_Simple(t *testing.T) {
	f := NewJSONFormatter()
	var buf bytes.Buffer

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	rctx := &RecordContext{Record: record}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if result["level"] != "info" {
		t.Errorf("level = %v, want %v", result["level"], "info")
	}
	if result["message"] != "test message" {
		t.Errorf("message = %v, want %v", result["message"], "test message")
	}
}

func TestJSONFormatter_Write_NonVerbose(t *testing.T) {
	f := NewJSONFormatter()
	// Default is non-verbose
	var buf bytes.Buffer

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	record.AddAttrs(slog.String("key", "value"))
	rctx := &RecordContext{Record: record}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Non-verbose should only have level and message
	if len(result) != 2 {
		t.Errorf("result length = %d, want 2", len(result))
	}
	if _, ok := result["key"]; ok {
		t.Error("non-verbose output contains attribute 'key'")
	}
}

func TestJSONFormatter_Write_AllLevels(t *testing.T) {
	levels := []struct {
		level slog.Level
		want  string
	}{
		{slog.LevelDebug, "debug"},
		{slog.LevelInfo, "info"},
		{slog.LevelWarn, "warn"},
		{slog.LevelError, "error"},
	}

	f := NewJSONFormatter()

	for _, tt := range levels {
		t.Run(tt.level.String(), func(t *testing.T) {
			var buf bytes.Buffer
			record := slog.NewRecord(time.Now(), tt.level, "test", 0)
			rctx := &RecordContext{Record: record}

			err := f.Write(&buf, rctx)
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			var result map[string]any
			if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if result["level"] != tt.want {
				t.Errorf("level = %v, want %v", result["level"], tt.want)
			}
		})
	}
}

func TestJSONFormatter_Write_NewlineTerminated(t *testing.T) {
	f := NewJSONFormatter()
	var buf bytes.Buffer

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	rctx := &RecordContext{Record: record}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got := buf.String()
	if !strings.HasSuffix(got, "\n") {
		t.Error("output does not end with newline")
	}
}

func TestResolveValue(t *testing.T) {
	tests := []struct {
		name  string
		value slog.Value
		check func(t *testing.T, result any)
	}{
		{
			name:  "string",
			value: slog.StringValue("test"),
			check: func(t *testing.T, result any) {
				if result != "test" {
					t.Errorf("result = %v, want test", result)
				}
			},
		},
		{
			name:  "int",
			value: slog.IntValue(42),
			check: func(t *testing.T, result any) {
				if result != int64(42) {
					t.Errorf("result = %v, want 42", result)
				}
			},
		},
		{
			name:  "bool",
			value: slog.BoolValue(true),
			check: func(t *testing.T, result any) {
				if result != true {
					t.Errorf("result = %v, want true", result)
				}
			},
		},
		{
			name: "group",
			value: slog.GroupValue(
				slog.String("a", "1"),
				slog.Int("b", 2),
			),
			check: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				if !ok {
					t.Fatal("result is not a map")
				}
				if m["a"] != "1" {
					t.Errorf("m[a] = %v, want 1", m["a"])
				}
				if m["b"] != int64(2) {
					t.Errorf("m[b] = %v, want 2", m["b"])
				}
			},
		},
		{
			name: "nested group",
			value: slog.GroupValue(
				slog.String("key", "value"),
				slog.Group("nested",
					slog.String("inner", "data"),
				),
			),
			check: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				if !ok {
					t.Fatal("result is not a map")
				}

				nested, ok := m["nested"].(map[string]any)
				if !ok {
					t.Fatal("nested is not a map")
				}

				if nested["inner"] != "data" {
					t.Errorf("nested[inner] = %v, want data", nested["inner"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveValue(tt.value)
			tt.check(t, result)
		})
	}
}

func TestJSONFormatter_SetVerbose(t *testing.T) {
	f := NewJSONFormatter()

	// Default should be non-verbose
	if f.Verbose() {
		t.Error("default verbose = true, want false")
	}

	// Test setting verbose
	f.SetVerbose(true)
	if !f.Verbose() {
		t.Error("after SetVerbose(true), Verbose() = false")
	}

	// Test chaining
	result := f.SetVerbose(false)
	if result != f {
		t.Error("SetVerbose() did not return formatter for chaining")
	}
	if f.Verbose() {
		t.Error("after SetVerbose(false), Verbose() = true")
	}
}

func TestJSONFormatter_Write_Verbose(t *testing.T) {
	f := NewJSONFormatter().SetVerbose(true)
	var buf bytes.Buffer

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	record.AddAttrs(
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
		slog.Bool("key3", true),
	)
	rctx := &RecordContext{Record: record}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verbose should include attributes
	if result["key1"] != "value1" {
		t.Errorf("key1 = %v, want value1", result["key1"])
	}
	if result["key2"] != float64(42) {
		t.Errorf("key2 = %v, want 42", result["key2"])
	}
	if result["key3"] != true {
		t.Errorf("key3 = %v, want true", result["key3"])
	}
}

func TestJSONFormatter_Write_VerboseWithGroups(t *testing.T) {
	f := NewJSONFormatter().SetVerbose(true)
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

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verbose should include groups
	groups, ok := result["groups"].([]any)
	if !ok {
		t.Fatal("groups field missing or not an array")
	}
	if len(groups) != 2 {
		t.Errorf("groups length = %d, want 2", len(groups))
	}
	if groups[0] != "group1" || groups[1] != "group2" {
		t.Errorf("groups = %v, want [group1 group2]", groups)
	}
}

func TestJSONFormatter_Write_VerboseWithNestedGroups(t *testing.T) {
	f := NewJSONFormatter().SetVerbose(true)
	var buf bytes.Buffer

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	record.AddAttrs(
		slog.String("key", "value"),
		slog.Group("outer",
			slog.String("a", "1"),
			slog.Group("inner",
				slog.String("b", "2"),
			),
		),
	)
	rctx := &RecordContext{Record: record}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Check nested structure
	outer, ok := result["outer"].(map[string]any)
	if !ok {
		t.Fatal("outer is not a map")
	}
	if outer["a"] != "1" {
		t.Errorf("outer[a] = %v, want 1", outer["a"])
	}

	inner, ok := outer["inner"].(map[string]any)
	if !ok {
		t.Fatal("inner is not a map")
	}
	if inner["b"] != "2" {
		t.Errorf("inner[b] = %v, want 2", inner["b"])
	}
}

func TestJSONFormatter_Write_InvalidJSON(t *testing.T) {
	f := NewJSONFormatter()
	var buf bytes.Buffer

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	// Add various types
	record.AddAttrs(
		slog.String("string", "value"),
		slog.Int("int", 123),
		slog.Float64("float", 1.23),
		slog.Bool("bool", true),
		slog.Duration("duration", 5*time.Second),
	)
	rctx := &RecordContext{Record: record}

	err := f.Write(&buf, rctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Verify it's valid JSON
	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
	}
}
