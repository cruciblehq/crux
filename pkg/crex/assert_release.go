//go:build !debug

package crex

// No-op assertions in release builds.
func Assert(condition bool, message string) {
	// No-op
}

// No-op assertions in release builds.
func Assertf(condition bool, format string, args ...any) {
	// No-op
}
