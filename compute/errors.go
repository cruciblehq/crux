package compute

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrUnknownProvider           = crex.New("unknown provider")
	ErrCruxd                     = crex.New("cruxd request failed")
	ErrIncompatibleCruxd         = crex.New("incompatible cruxd protocol version")
	ErrNotEnrolled               = crex.New("machine not enrolled")
	ErrMachineImageMissing       = crex.New("machine image not in cache")
	ErrMachineNotProvisioned     = crex.New("machine has not been provisioned")
	ErrMachineAlreadyProvisioned = crex.New("machine is already provisioned")
	ErrMachineNotRunning         = crex.New("machine is not running")
)
