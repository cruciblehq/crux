package mac

import "github.com/cruciblehq/crux/crex"

var (
	ErrCompile         = crex.New("MAC compile failed")
	ErrInvalidMAC      = crex.New("invalid MAC spec")
	ErrInvalidMACAllow = crex.New("invalid MAC allow rule")
	ErrInvalidMACExpr  = crex.New("invalid MAC expression")
	ErrInvalidMACValue = crex.New("invalid MAC value")
)
