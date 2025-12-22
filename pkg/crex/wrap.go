package crex

import "fmt"

// Wraps an underlying error with a sentinel error, creating an error chain.
//
// This function is intended for use in package-level (pkg/) code where errors
// need to be wrapped without adding user-facing context. It enforces the
// convention that the sentinel error comes first, followed by the underlying
// error, and both arguments must be errors (not strings).
func Wrap(sentinel error, err error) error {
	return fmt.Errorf("%w: %w", sentinel, err)
}
