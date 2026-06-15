package resource

import "github.com/cruciblehq/crux/crex"

var (
	ErrUnsupportedType      = crex.New("unsupported resource type")
	ErrBuildOutputNotFound  = crex.New("build output not found")
	ErrManifestNotFound     = crex.New("manifest not found")
	ErrResourceTypeMismatch = crex.New("resource type mismatch")
)
