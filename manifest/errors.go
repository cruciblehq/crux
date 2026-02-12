package manifest

import "errors"

var (
	ErrManifestReadFailed   = errors.New("failed to read manifest")
	ErrUnknownResourceType  = errors.New("unknown resource type")
	ErrInvalidCopyFormat    = errors.New("invalid copy format, expected 'src dest'")
	ErrInvalidFromFormat    = errors.New("invalid from format, expected 'file <path>' or '[ref] <identifier>'")
	ErrMissingFrom          = errors.New("from is required")
	ErrEmptyStep            = errors.New("step has no fields")
	ErrMultipleOperations   = errors.New("operations are mutually exclusive")
	ErrStepsWithoutPlatform = errors.New("steps requires platform")
	ErrOperationWithSteps   = errors.New("operations cannot have child steps")
	ErrShellWithCopy        = errors.New("shell cannot modify copy")
	ErrEnvWithCopy          = errors.New("env cannot modify copy")
	ErrUnknownPlatform      = errors.New("unknown platform")
	ErrNestedPlatform       = errors.New("platform cannot be set inside a platform group")
)
