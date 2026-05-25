package resource

import "errors"

var (
	ErrResolveHandler       = errors.New("failed to resolve handler")
	ErrBuildOutputNotFound  = errors.New("build output not found")
	ErrManifestNotFound     = errors.New("manifest not found")
	ErrResourceTypeMismatch = errors.New("resource type mismatch")
)
