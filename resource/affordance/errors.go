package affordance

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrResolution       = crex.New("affordance resolution failed")
	ErrUnknownSubsystem = crex.New("unknown subsystem")
)
