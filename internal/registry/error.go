package registry

import "fmt"

// Error response from the registry API.
//
// Provides both machine-readable error classification through the Code field
// and context through the Message field. The media type is [MediaTypeError].
type Error struct {

	// Error code (see [ErrorCode]).
	Code ErrorCode `codec:"code"`

	// Error description.
	Message string `codec:"message"`
}

// Implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Validates the error response.
//
// The code must be a known [ErrorCode] and the message must not be empty.
func (e *Error) Validate() error {
	if !isValidErrorCode(e.Code) {
		return ErrErrorCodeInvalid
	}
	if e.Message == "" {
		return ErrErrorMessageEmpty
	}
	return nil
}
