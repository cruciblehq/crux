package cli

import "errors"

var (
	ErrProviderNotFound    = errors.New("provider not found")
	ErrProviderUnsupported = errors.New("provider not supported")
)
