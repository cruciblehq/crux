package resource

import "errors"

var (
	ErrUnsupportedType      = errors.New("unsupported resource type")
	ErrBuildOutputNotFound  = errors.New("build output not found")
	ErrManifestNotFound     = errors.New("manifest not found")
	ErrResourceTypeMismatch = errors.New("resource type mismatch")
)
