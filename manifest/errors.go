package manifest

import "errors"

var (

	// Read errors.
	ErrManifestReadFailed  = errors.New("failed to read manifest")
	ErrUnknownResourceType = errors.New("unknown resource type")

	// Recipe errors.
	ErrMissingStages        = errors.New("at least one stage is required")
	ErrMissingFrom          = errors.New("from is required")
	ErrNoOutputStage        = errors.New("at least one stage must be non-transient")
	ErrMultipleOutputStages = errors.New("only one stage can be non-transient")
	ErrDuplicateStageName   = errors.New("duplicate stage name")

	// Step errors.
	ErrEmptyStep            = errors.New("step has no fields")
	ErrMultipleOperations   = errors.New("operations are mutually exclusive")
	ErrOperationWithSteps   = errors.New("operations cannot have child steps")
	ErrInvalidCopyFormat    = errors.New("invalid copy format, expected 'src dest'")
	ErrShellWithCopy        = errors.New("shell cannot modify copy")
	ErrEnvWithCopy          = errors.New("env cannot modify copy")
	ErrStepsWithoutPlatform = errors.New("steps requires platform")
	ErrUnknownPlatform      = errors.New("unknown platform")
	ErrNestedPlatform       = errors.New("platform cannot be set inside a platform group")
)
