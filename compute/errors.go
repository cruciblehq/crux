package compute

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrUnknownProvider   = crex.New("unknown provider")
	ErrStorage           = crex.New("storage operation failed")
	ErrNetwork           = crex.New("network operation failed")
	ErrCruxd             = crex.New("cruxd request failed")
	ErrIncompatibleCruxd = crex.New("incompatible cruxd protocol version")
)
