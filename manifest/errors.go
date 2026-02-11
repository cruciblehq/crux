package manifest

import "errors"

var (
	ErrManifestReadFailed   = errors.New("failed to read manifest")
	ErrUnknownResourceType  = errors.New("unknown resource type")
	ErrInvalidCopyFormat    = errors.New("invalid copy format, expected 'src dest'")
	ErrInvalidFromFormat    = errors.New("invalid from format, expected 'file <path>' or '[ref] <identifier>'")
	ErrMissingFrom          = errors.New("from is required")
	ErrEmptyStep            = errors.New("step has no fields")
	ErrMultipleOperations   = errors.New("run, exec, and copy are mutually exclusive")
	ErrStepsWithoutPlatform = errors.New("steps requires platform")
	ErrOperationWithSteps   = errors.New("operations cannot have child steps")
	ErrShellWithExec        = errors.New("shell cannot modify exec")
	ErrShellWithCopy        = errors.New("shell cannot modify copy")
	ErrEnvWithCopy          = errors.New("env cannot modify copy")
	ErrExecEmpty            = errors.New("exec requires at least one element")
	ErrUnknownPlatform      = errors.New("unknown platform")
	ErrNestedPlatform       = errors.New("platform cannot be set inside a platform group")
)
