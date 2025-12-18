package types

import (
	"errors"
)

var (
	ErrInvalidContentType     = errors.New("invalid content type")
	ErrUnsupportedContentType = errors.New("unsupported content type")
	ErrEncodingFailed         = errors.New("encoding failed")
	ErrDecodingFailed         = errors.New("decoding failed")
)
