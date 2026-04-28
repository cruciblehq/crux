package cap

import "github.com/cruciblehq/crex"

// Selects which capability sets a grant populates.
//
// The kernel maintains five capability sets per task: effective, permitted,
// inheritable, bounding, and ambient. Each mode maps to a specific combination
// of those sets, matching the Linux semantics for how capabilities propagate
// across fork and exec.
type Mode string

const (
	ModeFull        Mode = "full"        // All five sets. Broadest grant.
	ModeEffective   Mode = "effective"   // Effective, permitted, and bounding (effective immediately, survives exec, does not auto-inherit).
	ModeInheritable Mode = "inheritable" // Permitted, inheritable, ambient, and bounding (auto-inherits across exec via ambient).
	ModePermitted   Mode = "permitted"   // Permitted and bounding (raisable on demand, not effective by default).
	ModeBound       Mode = "bound"       // Bounding only (exec ceiling for child processes).
)

// Parses a mode string.
func ParseMode(s string) (Mode, error) {
	v := Mode(s)
	switch v {
	case ModeFull, ModeEffective, ModeInheritable, ModePermitted, ModeBound:
		return v, nil
	default:
		return "", crex.Wrapf(ErrInvalidGrant, "unknown mode %q", s)
	}
}
