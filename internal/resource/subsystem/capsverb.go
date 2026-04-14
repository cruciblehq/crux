package subsystem

import "github.com/cruciblehq/crex"

// Cap permission verb.
type capVerb string

const (
	capVerbGrant       capVerb = "grant"       // All five sets (effective immediately, survives exec, auto-inherits). Broadest grant.
	capVerbEffective   capVerb = "effective"   // Effective + permitted + bounding (effective immediately, survives exec, does not auto-inherit).
	capVerbInheritable capVerb = "inheritable" // Permitted + inheritable + ambient + bounding (auto-inherits across exec via ambient).
	capVerbPermitted   capVerb = "permitted"   // Permitted + bounding (raisable on demand, not effective by default).
	capVerbBound       capVerb = "bound"       // Bounding only (exec ceiling for child processes).
)

// Converts a string to a capVerb, returning an error for unknown values.
func parseCapVerb(s string) (capVerb, error) {
	v := capVerb(s)
	switch v {
	case capVerbGrant, capVerbEffective, capVerbInheritable, capVerbPermitted, capVerbBound:
		return v, nil
	default:
		return "", crex.Wrapf(ErrGrantExpression, "unknown verb %q", s)
	}
}
