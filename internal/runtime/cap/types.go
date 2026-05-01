package cap

import "github.com/cruciblehq/crux/internal/crex"

// Selects which capability sets a grant populates.
//
// The kernel maintains five capability sets per task: effective, permitted,
// inheritable, bounding, and ambient. Each mode maps to a specific combination
// of those sets, matching the Linux semantics for how capabilities propagate
// across fork and exec.
type mode string

const (
	modeFull        mode = "full"        // All five sets. Broadest grant.
	modeEffective   mode = "effective"   // Effective, permitted, and bounding (effective immediately, survives exec, does not auto-inherit).
	modeInheritable mode = "inheritable" // Permitted, inheritable, ambient, and bounding (auto-inherits across exec via ambient).
	modePermitted   mode = "permitted"   // Permitted and bounding (raisable on demand, not effective by default).
	modeBound       mode = "bound"       // Bounding only (exec ceiling for child processes).
)

// Parses a mode string.
func parseMode(s string) (mode, error) {
	v := mode(s)
	switch v {
	case modeFull, modeEffective, modeInheritable, modePermitted, modeBound:
		return v, nil
	default:
		return "", crex.Wrapf(ErrInvalidGrant, "unknown mode %q", s)
	}
}
