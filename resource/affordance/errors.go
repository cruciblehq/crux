package affordance

import "errors"

var (
	ErrResolution       = errors.New("affordance resolution failed")
	ErrUnknownSubsystem = errors.New("unknown subsystem")
	ErrConflict         = errors.New("duplicate grant")
)
