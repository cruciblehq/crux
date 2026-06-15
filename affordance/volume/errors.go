package volume

import "errors"

var (
	ErrInvalidGrant = errors.New("invalid volume grant")
	ErrInvalidMount = errors.New("invalid volume mount")
)
