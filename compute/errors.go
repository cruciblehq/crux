package compute

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrUnknownProvider   = crex.New("unknown provider")
	ErrCruxd             = crex.New("cruxd request failed")
	ErrIncompatibleCruxd = crex.New("incompatible cruxd protocol version")
	ErrNotEnrolled       = crex.New("machine not enrolled")
)
