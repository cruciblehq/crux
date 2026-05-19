package cap

import "errors"

var (
	ErrInvalidGrant = errors.New("invalid capability grant")
	ErrConflict     = errors.New("capability conflict")
)
