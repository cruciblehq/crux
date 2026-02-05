package crex

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
)

func TestError_Description(t *testing.T) {
	err := &Error{description: "test description"}
	if got := err.Description(); got != "test description" {
		t.Errorf("Description() = %q, want %q", got, "test description")
	}
}

func TestError_Reason(t *testing.T) {
	err := &Error{reason: "test reason"}
	if got := err.Reason(); got != "test reason" {
		t.Errorf("Reason() = %q, want %q", got, "test reason")
	}
}

func TestError_Fallback(t *testing.T) {
	err := &Error{fallback: "test fallback"}
	if got := err.Fallback(); got != "test fallback" {
		t.Errorf("Fallback() = %q, want %q", got, "test fallback")
	}
}

func TestError_Cause(t *testing.T) {
	cause := errors.New("underlying error")
	err := &Error{cause: cause}
	if got := err.Cause(); got != cause {
		t.Errorf("Cause() = %v, want %v", got, cause)
	}
}

func TestError_Class(t *testing.T) {
	err := &Error{class: ErrorClassUser}
	if got := err.Class(); got != ErrorClassUser {
		t.Errorf("Class() = %q, want %q", got, ErrorClassUser)
	}
}

func TestError_Context(t *testing.T) {
	t.Run("with context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "key", "value")
		err := &Error{context: ctx}
		if got := err.Context(); got != ctx {
			t.Errorf("Context() = %v, want %v", got, ctx)
		}
	})

	t.Run("without context", func(t *testing.T) {
		err := &Error{}
		if got := err.Context(); got != nil {
			t.Errorf("Context() = %v, want nil", got)
		}
	})
}

func TestError_Detail(t *testing.T) {
	err := &Error{
		details: map[string]any{
			"key1": "value1",
			"key2": 42,
		},
	}

	t.Run("existing key", func(t *testing.T) {
		val, ok := err.Detail("key1")
		if !ok {
			t.Error("Detail() ok = false, want true")
		}
		if val != "value1" {
			t.Errorf("Detail() = %v, want %v", val, "value1")
		}
	})

	t.Run("non-existing key", func(t *testing.T) {
		_, ok := err.Detail("nonexistent")
		if ok {
			t.Error("Detail() ok = true, want false")
		}
	})

	t.Run("nil details", func(t *testing.T) {
		err := &Error{}
		_, ok := err.Detail("key")
		if ok {
			t.Error("Detail() ok = true, want false for nil details")
		}
	})
}

func TestError_Details(t *testing.T) {
	t.Run("with details", func(t *testing.T) {
		original := map[string]any{
			"key1": "value1",
			"key2": 42,
		}
		err := &Error{details: original}
		details := err.Details()

		if len(details) != 2 {
			t.Errorf("Details() length = %d, want 2", len(details))
		}

		// Verify it's a copy
		details["key3"] = "new value"
		if _, ok := err.details["key3"]; ok {
			t.Error("Details() returned original map, want copy")
		}
	})

	t.Run("nil details", func(t *testing.T) {
		err := &Error{}
		details := err.Details()
		if details == nil {
			t.Error("Details() = nil, want empty map")
		}
		if len(details) != 0 {
			t.Errorf("Details() length = %d, want 0", len(details))
		}
	})
}

func TestError_DetailKeys(t *testing.T) {
	t.Run("with details", func(t *testing.T) {
		err := &Error{
			details: map[string]any{
				"zebra": 1,
				"alpha": 2,
				"beta":  3,
			},
		}
		keys := err.DetailKeys()
		want := []string{"alpha", "beta", "zebra"}

		if len(keys) != len(want) {
			t.Errorf("DetailKeys() length = %d, want %d", len(keys), len(want))
		}
		for i, key := range keys {
			if key != want[i] {
				t.Errorf("DetailKeys()[%d] = %q, want %q", i, key, want[i])
			}
		}
	})

	t.Run("empty details", func(t *testing.T) {
		err := &Error{}
		keys := err.DetailKeys()
		if len(keys) != 0 {
			t.Errorf("DetailKeys() length = %d, want 0", len(keys))
		}
	})
}

func TestError_Error(t *testing.T) {
	err := &Error{
		description: "failed",
		reason:      "bad input",
	}
	if got := err.Error(); got != "failed: bad input" {
		t.Errorf("Error() = %q, want %q", got, "failed: bad input")
	}
}

func TestError_Unwrap(t *testing.T) {
	cause := errors.New("cause")
	err := &Error{cause: cause}
	if got := err.Unwrap(); got != cause {
		t.Errorf("Unwrap() = %v, want %v", got, cause)
	}

	// Test with errors.Is
	if !errors.Is(err, cause) {
		t.Error("errors.Is() = false, want true")
	}
}

func TestError_String(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "all fields",
			err: &Error{
				description: "operation failed",
				reason:      "invalid input",
				fallback:    "Use valid input",
			},
			want: "operation failed: invalid input. Use valid input",
		},
		{
			name: "no reason",
			err: &Error{
				description: "operation failed",
				fallback:    "Try again",
			},
			want: "operation failed. Try again",
		},
		{
			name: "no fallback",
			err: &Error{
				description: "operation failed",
				reason:      "invalid input",
			},
			want: "operation failed: invalid input",
		},
		{
			name: "only description",
			err: &Error{
				description: "operation failed",
			},
			want: "operation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestError_Format(t *testing.T) {
	err := &Error{
		description: "failed",
		reason:      "bad",
	}

	tests := []struct {
		name   string
		format string
		want   string
	}{
		{"v verb", "%v", "failed: bad"},
		{"s verb", "%s", "failed: bad"},
		{"q verb", "%q", `"failed: bad"`},
		{"unknown verb", "%x", "failed: bad"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fmt.Sprintf(tt.format, err)
			if got != tt.want {
				t.Errorf("Format(%s) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}

func TestError_LogValue(t *testing.T) {
	err := &Error{
		class:       ErrorClassUser,
		description: "test failed",
		reason:      "bad input",
		fallback:    "fix it",
		cause:       errors.New("underlying"),
		details: map[string]any{
			"key": "value",
		},
	}

	val := err.LogValue()
	if val.Kind() != slog.KindGroup {
		t.Errorf("LogValue().Kind() = %v, want KindGroup", val.Kind())
	}

	// Verify required attributes
	attrMap := make(map[string]slog.Attr)
	for _, attr := range val.Group() {
		attrMap[attr.Key] = attr
	}

	if attrMap["class"].Value.String() != "user" {
		t.Errorf("class = %q, want %q", attrMap["class"].Value.String(), "user")
	}
	if attrMap["description"].Value.String() != "test failed" {
		t.Errorf("description = %q, want %q", attrMap["description"].Value.String(), "test failed")
	}
}
