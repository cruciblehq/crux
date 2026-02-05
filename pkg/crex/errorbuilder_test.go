package crex

import (
	"context"
	"errors"
	"testing"
)

func TestUserError(t *testing.T) {
	err := UserError("test", "reason").Err()
	crexErr := err.(*Error)

	if crexErr.Class() != ErrorClassUser {
		t.Errorf("Class() = %v, want %v", crexErr.Class(), ErrorClassUser)
	}
	if crexErr.Description() != "test" {
		t.Errorf("Description() = %q, want %q", crexErr.Description(), "test")
	}
	if crexErr.Reason() != "reason" {
		t.Errorf("Reason() = %q, want %q", crexErr.Reason(), "reason")
	}
}

func TestUserErrorf(t *testing.T) {
	err := UserErrorf("test", "value is %d", 42).Err()
	crexErr := err.(*Error)

	if crexErr.Reason() != "value is 42" {
		t.Errorf("Reason() = %q, want %q", crexErr.Reason(), "value is 42")
	}
}

func TestSystemError(t *testing.T) {
	err := SystemError("test", "reason").Err()
	crexErr := err.(*Error)

	if crexErr.Class() != ErrorClassSystem {
		t.Errorf("Class() = %v, want %v", crexErr.Class(), ErrorClassSystem)
	}
}

func TestSystemErrorf(t *testing.T) {
	err := SystemErrorf("test", "port %d unavailable", 8080).Err()
	crexErr := err.(*Error)

	if crexErr.Reason() != "port 8080 unavailable" {
		t.Errorf("Reason() = %q, want %q", crexErr.Reason(), "port 8080 unavailable")
	}
}

func TestProgrammingError(t *testing.T) {
	err := ProgrammingError("test", "reason").Err()
	crexErr := err.(*Error)

	if crexErr.Class() != ErrorClassProgramming {
		t.Errorf("Class() = %v, want %v", crexErr.Class(), ErrorClassProgramming)
	}
}

func TestProgrammingErrorf(t *testing.T) {
	err := ProgrammingErrorf("test", "index %d out of bounds", 5).Err()
	crexErr := err.(*Error)

	if crexErr.Reason() != "index 5 out of bounds" {
		t.Errorf("Reason() = %q, want %q", crexErr.Reason(), "index 5 out of bounds")
	}
}

func TestBug(t *testing.T) {
	err := Bug("test", "reason").Err()
	crexErr := err.(*Error)

	if crexErr.Class() != ErrorClassProgramming {
		t.Errorf("Class() = %v, want %v", crexErr.Class(), ErrorClassProgramming)
	}
}

func TestBugf(t *testing.T) {
	err := Bugf("test", "nil pointer at line %d", 123).Err()
	crexErr := err.(*Error)

	if crexErr.Reason() != "nil pointer at line 123" {
		t.Errorf("Reason() = %q, want %q", crexErr.Reason(), "nil pointer at line 123")
	}
}

func TestNewError_EmptyDescription_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("newError() with empty description did not panic")
		}
	}()

	newError(ErrorClassUser, "", "reason")
}

func TestNewError_EmptyReason_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("newError() with empty reason did not panic")
		}
	}()

	newError(ErrorClassUser, "description", "")
}

func TestNewError_WhitespaceOnly_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("newError() with whitespace-only description did not panic")
		}
	}()

	newError(ErrorClassUser, "  \t\n  ", "reason")
}

func TestNewError_TrimsSurroundingWhitespace(t *testing.T) {
	builder := newError(ErrorClassUser, "  test  ", "  reason  ")
	crexErr := builder.err

	if crexErr.description != "test" {
		t.Errorf("description = %q, want %q", crexErr.description, "test")
	}
	if crexErr.reason != "reason" {
		t.Errorf("reason = %q, want %q", crexErr.reason, "reason")
	}
}

func TestErrorBuilder_Fallback(t *testing.T) {
	err := UserError("test", "reason").
		Fallback("try again").
		Err()
	crexErr := err.(*Error)

	if crexErr.Fallback() != "try again" {
		t.Errorf("Fallback() = %q, want %q", crexErr.Fallback(), "try again")
	}
}

func TestErrorBuilder_Fallbackf(t *testing.T) {
	err := UserError("test", "reason").
		Fallbackf("retry in %d seconds", 30).
		Err()
	crexErr := err.(*Error)

	if crexErr.Fallback() != "retry in 30 seconds" {
		t.Errorf("Fallback() = %q, want %q", crexErr.Fallback(), "retry in 30 seconds")
	}
}

func TestErrorBuilder_Fallback_TrimsWhitespace(t *testing.T) {
	err := UserError("test", "reason").
		Fallback("  try again  ").
		Err()
	crexErr := err.(*Error)

	if crexErr.Fallback() != "try again" {
		t.Errorf("Fallback() = %q, want %q", crexErr.Fallback(), "try again")
	}
}

func TestErrorBuilder_Cause(t *testing.T) {
	underlying := errors.New("underlying error")
	err := UserError("test", "reason").
		Cause(underlying).
		Err()
	crexErr := err.(*Error)

	if crexErr.Cause() != underlying {
		t.Errorf("Cause() = %v, want %v", crexErr.Cause(), underlying)
	}

	// Verify errors.Is works
	if !errors.Is(err, underlying) {
		t.Error("errors.Is() = false, want true")
	}
}

func TestErrorBuilder_Detail(t *testing.T) {
	err := UserError("test", "reason").
		Detail("key1", "value1").
		Detail("key2", 42).
		Err()
	crexErr := err.(*Error)

	val1, ok1 := crexErr.Detail("key1")
	if !ok1 || val1 != "value1" {
		t.Errorf("Detail(key1) = (%v, %v), want (value1, true)", val1, ok1)
	}

	val2, ok2 := crexErr.Detail("key2")
	if !ok2 || val2 != 42 {
		t.Errorf("Detail(key2) = (%v, %v), want (42, true)", val2, ok2)
	}
}

func TestErrorBuilder_Context(t *testing.T) {
	ctx := context.WithValue(context.Background(), "key", "value")
	err := UserError("test", "reason").
		Context(ctx).
		Err()
	crexErr := err.(*Error)

	if crexErr.Context() != ctx {
		t.Errorf("Context() = %v, want %v", crexErr.Context(), ctx)
	}
}

func TestErrorBuilder_Validate(t *testing.T) {
	tests := []struct {
		name    string
		builder *ErrorBuilder
		wantErr bool
	}{
		{
			name:    "valid",
			builder: UserError("test", "reason"),
			wantErr: false,
		},
		{
			name: "empty description",
			builder: &ErrorBuilder{
				err: Error{description: "", reason: "reason"},
			},
			wantErr: true,
		},
		{
			name: "empty reason",
			builder: &ErrorBuilder{
				err: Error{description: "test", reason: ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.builder.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestErrorBuilder_Chaining(t *testing.T) {
	underlying := errors.New("underlying")
	ctx := context.Background()

	err := UserError("operation failed", "invalid input").
		Fallback("Use valid input").
		Cause(underlying).
		Detail("field", "username").
		Detail("value", "abc").
		Context(ctx).
		Err()

	crexErr := err.(*Error)

	if crexErr.Description() != "operation failed" {
		t.Errorf("Description() = %q, want %q", crexErr.Description(), "operation failed")
	}
	if crexErr.Reason() != "invalid input" {
		t.Errorf("Reason() = %q, want %q", crexErr.Reason(), "invalid input")
	}
	if crexErr.Fallback() != "Use valid input" {
		t.Errorf("Fallback() = %q, want %q", crexErr.Fallback(), "Use valid input")
	}
	if crexErr.Cause() != underlying {
		t.Errorf("Cause() = %v, want %v", crexErr.Cause(), underlying)
	}
	if crexErr.Context() != ctx {
		t.Errorf("Context() = %v, want %v", crexErr.Context(), ctx)
	}

	field, ok := crexErr.Detail("field")
	if !ok || field != "username" {
		t.Errorf("Detail(field) = (%v, %v), want (username, true)", field, ok)
	}
}
