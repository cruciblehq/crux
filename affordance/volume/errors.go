package volume

import "github.com/cruciblehq/crux/crex"

var (
	ErrInvalidGrant = crex.New("invalid volume grant")
	ErrInvalidMount = crex.New("invalid volume mount")
)
