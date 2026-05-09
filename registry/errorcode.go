package registry

// Platform-specific error code for machine-readable error classification.
//
// Provides granular error information beyond HTTP status codes, enabling
// clients to implement specific error handling logic. Used in the Code field
// of [Error] responses.
type ErrorCode string

const (
	ErrorCodeBadRequest           ErrorCode = "bad_request"                     // Request validation failed (malformed body, invalid fields).
	ErrorCodeNotFound             ErrorCode = "not_found"                       // Requested namespace, resource, version, or channel does not exist.
	ErrorCodeNamespaceExists      ErrorCode = "namespace_exists"                // Cannot create namespace: name already in use.
	ErrorCodeNamespaceNotEmpty    ErrorCode = "namespace_not_empty"             // Cannot delete namespace: contains resources.
	ErrorCodeResourceExists       ErrorCode = "resource_exists"                 // Cannot create resource: name already in use within namespace.
	ErrorCodeResourceHasPublished ErrorCode = "resource_has_published_versions" // Cannot delete resource: contains published versions.
	ErrorCodeVersionExists        ErrorCode = "version_exists"                  // Cannot create version: version string already in use.
	ErrorCodeVersionPublished     ErrorCode = "version_published"               // Cannot modify or delete version: already published and immutable.
	ErrorCodeChannelExists        ErrorCode = "channel_exists"                  // Cannot create channel: name already in use.
	ErrorCodePreconditionFailed   ErrorCode = "precondition_failed"             // Request precondition not met (e.g., If-Match header mismatch).
	ErrorCodeInternalError        ErrorCode = "internal_error"                  // Unexpected server error occurred.
)

// Set of known error codes for validation.
var validErrorCodes = map[ErrorCode]bool{
	ErrorCodeBadRequest:           true,
	ErrorCodeNotFound:             true,
	ErrorCodeNamespaceExists:      true,
	ErrorCodeNamespaceNotEmpty:    true,
	ErrorCodeResourceExists:       true,
	ErrorCodeResourceHasPublished: true,
	ErrorCodeVersionExists:        true,
	ErrorCodeVersionPublished:     true,
	ErrorCodeChannelExists:        true,
	ErrorCodePreconditionFailed:   true,
	ErrorCodeInternalError:        true,
}

// Whether the error code is a known value.
func isValidErrorCode(code ErrorCode) bool {
	return validErrorCodes[code]
}
