//go:build !debug

package crex

// Assert is a no-op in release builds.
func Assert(condition bool, message string) {
	// No-op.
}

// Assertf is a no-op in release builds.
func Assertf(condition bool, format string, args ...any) {
	// No-op.
}
