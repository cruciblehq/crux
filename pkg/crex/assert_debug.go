//go:build debug

package crex

import (
	"fmt"
	"runtime"
)

// Panics if the condition is false.
func Assert(condition bool, message string) {
	assert(condition, message)
}

// Panics if the condition is false.
func Assertf(condition bool, format string, args ...any) {
	assert(condition, fmt.Sprintf(format, args...))
}

// Internal helper for assertions.
func assert(condition bool, message string) {
	if !condition {
		_, file, line, _ := runtime.Caller(2)
		panic(fmt.Sprintf("assertion failed: %s\n  at %s:%d", message, file, line))
	}
}
