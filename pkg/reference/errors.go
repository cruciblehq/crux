package reference

import "errors"

var (
	ErrInvalidIdentifier = errors.New("invalid identifier")
	ErrInvalidReference  = errors.New("invalid reference")
	ErrInvalidDigest     = errors.New("invalid digest")
	ErrTypeMismatch      = errors.New("resource type mismatch")
)
