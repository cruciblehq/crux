package registry

import "errors"

var (

	// Validation errors.

	ErrNameEmpty         = errors.New("name cannot be empty")
	ErrNameTooLong       = errors.New("name cannot exceed 63 characters")
	ErrNameInvalid       = errors.New("name must contain only lowercase letters, numbers, and hyphens, and must start and end with an alphanumeric character")
	ErrVersionInvalid    = errors.New("version format must be semantic version")
	ErrTimestampInvalid  = errors.New("timestamp must be a positive unix epoch")
	ErrTimestampOrder    = errors.New("updatedAt must not be before createdAt")
	ErrTypeEmpty         = errors.New("resource type cannot be empty")
	ErrArchiveEmpty      = errors.New("archive URL cannot be empty when set")
	ErrSizeInvalid       = errors.New("archive size must be positive")
	ErrDigestInvalid     = errors.New("digest must be in algorithm:hex format")
	ErrArchiveIncomplete = errors.New("archive, size, and digest must all be set or all be null")
	ErrCountNegative     = errors.New("count must not be negative")
	ErrErrorCodeInvalid  = errors.New("error code must be a known value")
	ErrErrorMessageEmpty = errors.New("error message cannot be empty")

	// Type validation errors.

	ErrInvalidNamespace = errors.New("invalid namespace")
	ErrInvalidResource  = errors.New("invalid resource")
	ErrInvalidVersion   = errors.New("invalid version")
	ErrInvalidChannel   = errors.New("invalid channel")

	// Client errors.

	ErrMarshal          = errors.New("failed to marshal request body")
	ErrBaseURL          = errors.New("failed to parse base URL")
	ErrHTTPRequest      = errors.New("failed to create HTTP request")
	ErrHTTPExecute      = errors.New("failed to execute HTTP request")
	ErrHTTPStatus       = errors.New("unexpected HTTP status")
	ErrResponseDecode   = errors.New("failed to decode response body")
	ErrTypeMismatch     = errors.New("resource type mismatch")
	ErrNoVersions       = errors.New("no versions found")
	ErrNoMatchingVersion = errors.New("no matching version found")
)
