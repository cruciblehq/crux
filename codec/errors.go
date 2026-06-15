package codec

import "github.com/cruciblehq/crux/crex"

var (
	ErrUnsupportedFormat = crex.New("unsupported codec format")
	ErrNotValidatable    = crex.New("type does not implement Validatable")
	ErrInvalidInput      = crex.New("invalid input: must be a pointer to a struct")
	ErrMissingField      = crex.New("field not found on struct")
	ErrSetDefault        = crex.New("failed to set default value")
	ErrUnsupportedType   = crex.New("unsupported field type")
)
