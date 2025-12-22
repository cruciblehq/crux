package manifest

import "errors"

var (
	ErrManifestReadFailed  = errors.New("failed to read manifest")
	ErrUnknownResourceType = errors.New("unknown resource type")
)
