package crex

import (
	"strings"
	"testing"
)

func TestAssert_True(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Assert() panicked with true condition: %v", r)
		}
	}()

	Assert(true, "should not panic")
}

func TestAssertf_True(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Assertf() panicked with true condition: %v", r)
		}
	}()

	Assertf(true, "should not panic: %s", "test")
}

// Note: The following tests only work when compiled with -tags=debug
// In release builds, assertions are no-ops and will not panic

// Verifies that r is a string panic value containing every substring in want.
func checkAssertPanic(t *testing.T, r any, want ...string) {
	t.Helper()
	msg, ok := r.(string)
	if !ok {
		t.Errorf("panic value is not a string: %v", r)
		return
	}
	for _, sub := range want {
		if !strings.Contains(msg, sub) {
			t.Errorf("panic message does not contain %q: %s", sub, msg)
		}
	}
}

func TestAssert_False_Debug(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			checkAssertPanic(t, r, "assertion failed", "test message", "at ", "assert_test.go")
		} else {
			// In release builds, this is expected (no-op)
			t.Skip("Assert is no-op in release builds (compiled without -tags=debug)")
		}
	}()

	Assert(false, "test message")
}

func TestAssertf_False_Debug(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			checkAssertPanic(t, r, "assertion failed", "test value: 42", "at ")
		} else {
			// In release builds, this is expected (no-op)
			t.Skip("Assertf is no-op in release builds (compiled without -tags=debug)")
		}
	}()

	Assertf(false, "test value: %d", 42)
}
