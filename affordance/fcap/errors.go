package fcap

import "github.com/cruciblehq/crux/crex"

var (
	ErrInvalidGrant        = crex.New("invalid fcap grant")
	ErrUnknownFcapMode     = crex.New("unknown fcap mode")
	ErrInvalidFcap         = crex.New("invalid fcap spec")
	ErrInvalidCapabilities = crex.New("invalid fcap capabilities")
)
