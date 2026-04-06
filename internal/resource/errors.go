package resource

import "errors"

var (

	// General.

	ErrBuild            = errors.New("build failed")
	ErrMissingOption    = errors.New("missing required option")
	ErrMissingRegistry  = errors.New("default registry is required")
	ErrMissingNamespace = errors.New("default namespace is required")
	ErrReadManifest     = errors.New("failed to read manifest")
	ErrResolveBuilder   = errors.New("failed to resolve builder")
	ErrUnsupported      = errors.New("unsupported operation")

	// File system.

	ErrFileSystemOperation = errors.New("file system operation failed")
	ErrInvalidPath         = errors.New("invalid path")

	// Cache.

	ErrCacheOperation = errors.New("cache operation failed")

	// Blueprint.

	ErrBlueprintBuild = errors.New("blueprint build failed")

	// Affordance.

	ErrSourceResolve    = errors.New("failed to resolve image source")
	ErrResolutionCycle  = errors.New("affordance resolution cycle")
	ErrResolutionFailed = errors.New("affordance resolution failed")
)
