package fcap

import "errors"

var (
	ErrInvalidGrant        = errors.New("invalid fcap grant")
	ErrUnknownFcapMode     = errors.New("unknown fcap mode")
	ErrInvalidFcap         = errors.New("invalid fcap spec")
	ErrInvalidCapabilities = errors.New("invalid fcap capabilities")
)
