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

// Parses a mode string, returning an error for unknown values.
func ParseMode(s string) (Mode, error) {
	v := Mode(s)
	switch v {
	case ModeFull, ModeEffective, ModeInheritable, ModePermitted, ModeBound:
		return v, nil
	default:
		return "", crex.Wrapf(ErrInvalidRule, "unknown mode %q", s)
	}
}

// Rule expression for capability grants.
//
// Each rule names a capability and the mode that selects which of the
// five kernel capability sets to populate.
type Grant struct {
	Mode Mode   // Which capability sets to populate.
	Name string // Capability name (without the CAP_ prefix, e.g. "net_admin").
}
