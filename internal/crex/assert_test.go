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

func TestAssert_False_Debug(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			msg, ok := r.(string)
			if !ok {
				t.Errorf("panic value is not a string: %v", r)
				return
			}

			if !strings.Contains(msg, "assertion failed") {
				t.Errorf("panic message does not contain 'assertion failed': %s", msg)
			}
			if !strings.Contains(msg, "test message") {
				t.Errorf("panic message does not contain 'test message': %s", msg)
			}
			if !strings.Contains(msg, "at ") {
				t.Errorf("panic message does not contain file location: %s", msg)
			}
			if !strings.Contains(msg, "assert_test.go") {
				t.Errorf("panic message does not contain filename: %s", msg)
			}
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
			msg, ok := r.(string)
			if !ok {
				t.Errorf("panic value is not a string: %v", r)
				return
			}

			if !strings.Contains(msg, "assertion failed") {
				t.Errorf("panic message does not contain 'assertion failed': %s", msg)
			}
			if !strings.Contains(msg, "test value: 42") {
				t.Errorf("panic message does not contain formatted text: %s", msg)
			}
			if !strings.Contains(msg, "at ") {
				t.Errorf("panic message does not contain file location: %s", msg)
			}
		} else {
			// In release builds, this is expected (no-op)
			t.Skip("Assertf is no-op in release builds (compiled without -tags=debug)")
		}
	}()

	Assertf(false, "test value: %d", 42)
}
