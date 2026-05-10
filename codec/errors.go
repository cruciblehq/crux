package codec

import "errors"

var (
	ErrUnsupportedFormat = errors.New("unsupported codec format")
	ErrNotValidatable    = errors.New("type does not implement Validatable")
	ErrInvalidInput      = errors.New("invalid input: must be a pointer to a struct")
	ErrMissingField      = errors.New("field not found on struct")
)
