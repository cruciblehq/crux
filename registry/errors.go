package registry

import "errors"

var (
	ErrNoVersions        = errors.New("no versions found")
	ErrNoMatchingVersion = errors.New("no matching version")
	ErrTypeMismatch      = errors.New("resource type mismatch")

	// Validation errors.
	ErrNameEmpty          = errors.New("name cannot be empty")
	ErrNameTooLong        = errors.New("name cannot exceed 63 characters")
	ErrNameInvalid        = errors.New("name must contain only lowercase letters, numbers, and hyphens, and must start and end with an alphanumeric character")
	ErrVersionInvalid     = errors.New("invalid version format: must be semantic version (e.g., 1.2.3, 1.0.0-alpha.1)")

	// Client errors.
	ErrMarshal            = errors.New("failed to marshal request body")
	ErrHTTPRequest        = errors.New("failed to create HTTP request")
	ErrHTTPExecute        = errors.New("failed to execute HTTP request")
	ErrHTTPStatus         = errors.New("unexpected HTTP status")
	ErrResponseDecode     = errors.New("failed to decode response")
	ErrBaseURL            = errors.New("failed to parse base URL")
)
