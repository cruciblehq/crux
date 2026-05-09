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

// Wraps a sentinel error with a formatted message, creating an error chain.
//
// This is a convenience for the common pattern of wrapping a sentinel with
// an inline fmt.Errorf call:
//
//	crex.Wrapf(ErrBuild, "stage %s: %w", label, err)
//
// is equivalent to:
//
//	crex.Wrap(ErrBuild, fmt.Errorf("stage %s: %w", label, err))
func Wrapf(sentinel error, format string, args ...any) error {
	return fmt.Errorf("%w: %w", sentinel, fmt.Errorf(format, args...))
}
