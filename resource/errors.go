package resource

import "errors"

var (
	ErrResolveHandler       = errors.New("failed to resolve handler")
	ErrBuildOutputNotFound  = errors.New("build output not found")
	ErrManifestNotFound     = errors.New("manifest not found")
	ErrResourceTypeMismatch = errors.New("resource type mismatch")
)

var (
	ErrBuildPlan           = errors.New("blueprint build failed")
	ErrBuild               = errors.New("build failed")
	ErrFileSystemOperation = errors.New("file system operation failed")
	ErrInvalidPath         = errors.New("invalid build path")
	ErrBuildWidget         = errors.New("build failed")
	ErrResolution          = errors.New("affordance resolution failed")
	ErrUnknownSubsystem    = errors.New("unknown subsystem")
	ErrConflict            = errors.New("duplicate grant")
)
