package registry

import "errors"

var (

	// Resolve errors.
	ErrNoVersions        = errors.New("no versions found")
	ErrNoMatchingVersion = errors.New("no matching version")
	ErrTypeMismatch      = errors.New("resource type mismatch")

	// Client errors.
	ErrMarshal        = errors.New("failed to marshal request body")
	ErrHTTPRequest    = errors.New("failed to create HTTP request")
	ErrHTTPExecute    = errors.New("failed to execute HTTP request")
	ErrHTTPStatus     = errors.New("unexpected HTTP status")
	ErrResponseDecode = errors.New("failed to decode response")
	ErrBaseURL        = errors.New("failed to parse base URL")
)
