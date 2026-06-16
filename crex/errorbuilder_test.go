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

func TestErrorBuilder_Recovery(t *testing.T) {
	err := UserError("test", "reason").
		Recovery("try again").
		Err()
	crexErr := err.(*Error)

	if crexErr.Recovery() != "try again" {
		t.Errorf("Recovery() = %q, want %q", crexErr.Recovery(), "try again")
	}
}

func TestErrorBuilder_Recoveryf(t *testing.T) {
	err := UserError("test", "reason").
		Recoveryf("retry in %d seconds", 30).
		Err()
	crexErr := err.(*Error)

	if crexErr.Recovery() != "retry in 30 seconds" {
		t.Errorf("Recovery() = %q, want %q", crexErr.Recovery(), "retry in 30 seconds")
	}
}

func TestErrorBuilder_Recovery_TrimsWhitespace(t *testing.T) {
	err := UserError("test", "reason").
		Recovery("  try again  ").
		Err()
	crexErr := err.(*Error)

	if crexErr.Recovery() != "try again" {
		t.Errorf("Recovery() = %q, want %q", crexErr.Recovery(), "try again")
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

func TestErrorBuilder_Chaining(t *testing.T) {
	underlying := errors.New("underlying")
	ctx := context.Background()

	err := UserError("operation failed", "invalid input").
		Recovery("Use valid input").
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
	if crexErr.Recovery() != "Use valid input" {
		t.Errorf("Recovery() = %q, want %q", crexErr.Recovery(), "Use valid input")
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

func TestErrorBuilder_Ensure_ClasslessGetsClassed(t *testing.T) {
	cause := errors.New("boom")
	err := SystemError("fetch failed", "registry unreachable").Ensure(cause)

	class, ok := ClassOf(err)
	if !ok || class != ErrorClassSystem {
		t.Errorf("ClassOf() = (%v, %v), want (system, true)", class, ok)
	}
	if !errors.Is(err, cause) {
		t.Error("errors.Is(err, cause) = false, want true")
	}
}

func TestErrorBuilder_Ensure_RespectsExistingClass(t *testing.T) {
	inner := UserError("bad input", "missing flag").Err()
	err := SystemError("fetch failed", "registry unreachable").Ensure(inner)

	if err != inner {
		t.Error("Ensure() replaced an already-classed error, want no-op")
	}
	class, _ := ClassOf(err)
	if class != ErrorClassUser {
		t.Errorf("ClassOf() = %v, want user (first classification wins)", class)
	}
}

func TestErrorBuilder_Ensure_NilReturnsNil(t *testing.T) {
	if got := UserError("x", "y").Ensure(nil); got != nil {
		t.Errorf("Ensure(nil) = %v, want nil", got)
	}
}

func TestErrorBuilder_Reclassify_OverridesExistingClass(t *testing.T) {
	inner := UserError("bad input", "missing flag").Err()
	err := SystemError("fetch failed", "registry unreachable").Reclassify(inner)

	class, ok := ClassOf(err)
	if !ok || class != ErrorClassSystem {
		t.Errorf("ClassOf() = (%v, %v), want (system, true)", class, ok)
	}
	// The prior classification stays in the chain for errors.Is.
	if !errors.Is(err, inner) {
		t.Error("errors.Is(err, inner) = false, want true")
	}
}

func TestErrorBuilder_Reclassify_NilReturnsNil(t *testing.T) {
	if got := UserError("x", "y").Reclassify(nil); got != nil {
		t.Errorf("Reclassify(nil) = %v, want nil", got)
	}
}

func TestClassOf_Unclassified(t *testing.T) {
	class, ok := ClassOf(errors.New("plain"))
	if ok || class != ErrorClassUnknown {
		t.Errorf("ClassOf(plain) = (%v, %v), want (unknown, false)", class, ok)
	}
}
