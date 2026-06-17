package affordance

import "github.com/cruciblehq/crux/crex"

var (
	ErrResolution       = crex.New("affordance resolution failed")
	ErrUnknownSubsystem = crex.New("unknown subsystem")
)
