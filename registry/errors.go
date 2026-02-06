package registry

import "errors"

// Note: error management is still pretty bogus on this package, relying on a
// mix of sentinel errors and error wrapping. This should be cleaned up.

var (
	ErrNoVersions        = errors.New("no versions found")
	ErrNoMatchingVersion = errors.New("no matching version")
	ErrTypeMismatch      = errors.New("resource type mismatch")
)
