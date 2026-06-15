package mac

import "errors"

var (
	ErrCompile         = errors.New("MAC compile failed")
	ErrInvalidMAC      = errors.New("invalid MAC spec")
	ErrInvalidMACAllow = errors.New("invalid MAC allow rule")
	ErrInvalidMACExpr  = errors.New("invalid MAC expression")
	ErrInvalidMACValue = errors.New("invalid MAC value")
)
