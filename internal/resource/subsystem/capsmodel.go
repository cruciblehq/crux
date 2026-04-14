package subsystem

import "github.com/cruciblehq/crex"

// Controls which privileged kernel operations are allowed to the container.
//
// Linux capabilities divide root privilege into individual units (e.g., CHOWN,
// NET_BIND_SERVICE, SYS_PTRACE). Each field holds the names of capabilities
// granted to one of the five kernel sets: effective, permitted, inheritable,
// bounding, and ambient. The zero value grants no capabilities. Use the Grant
// methods to mutate; they enforce the kernel invariants between sets.
type caps struct {
	Effective   []string `codec:"effective,omitempty"`   // Effective capability set.
	Permitted   []string `codec:"permitted,omitempty"`   // Permitted capability set.
	Inheritable []string `codec:"inheritable,omitempty"` // Inheritable capability set.
	Bounding    []string `codec:"bounding,omitempty"`    // Bounding capability set.
	Ambient     []string `codec:"ambient,omitempty"`     // Ambient capability set.
}

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

// Grants a capability to all five sets.
//
// The capability is effective immediately, survives exec, and auto-inherits
// to child processes. This is the broadest grant.
func (c *caps) Grant(cap string) {
	appendUnique(&c.Effective, cap)
	appendUnique(&c.Permitted, cap)
	appendUnique(&c.Inheritable, cap)
	appendUnique(&c.Bounding, cap)
	appendUnique(&c.Ambient, cap)
}

// Grants a capability to the effective, permitted, and bounding sets.
//
// The capability is effective immediately and survives exec (via bounding),
// but does not auto-inherit to child processes. Useful for capabilities the
// service itself needs.
func (c *caps) GrantEffective(cap string) {
	appendUnique(&c.Effective, cap)
	appendUnique(&c.Permitted, cap)
	appendUnique(&c.Bounding, cap)
}

// Grants a capability that auto-inherits across exec.
//
// The capability is not effective in the current process, but after execve
// the ambient set automatically raises it into the child's effective and
// permitted sets. Useful for capabilities a service's children need but the
// parent doesn't use directly.
func (c *caps) GrantInheritable(cap string) {
	appendUnique(&c.Permitted, cap)
	appendUnique(&c.Inheritable, cap)
	appendUnique(&c.Ambient, cap)
	appendUnique(&c.Bounding, cap)
}

// Grants a capability to the permitted and bounding sets.
//
// The process may raise it into its effective set at will, and the bounding
// set allows it to persist across exec. Not effective by default, and does
// not auto-inherit. Useful for capabilities that are only needed for specific
// operations.
func (c *caps) GrantPermitted(cap string) {
	appendUnique(&c.Permitted, cap)
	appendUnique(&c.Bounding, cap)
}

// Grants a capability only in the bounding set.
//
// This acts as an exec ceiling: child processes may receive this capability
// (via file caps or ambient), but the current process cannot use it. Useful
// for capabilities that are only needed by child processes.
func (c *caps) GrantBound(cap string) {
	appendUnique(&c.Bounding, cap)
}

// Merges another capability set into this one.
//
// Returns true if any capability was added. Capabilities are purely additive,
// so no conflicts are possible.
func (c *caps) merge(other caps) bool {
	changed := mergeSlice(&c.Effective, other.Effective)
	changed = mergeSlice(&c.Permitted, other.Permitted) || changed
	changed = mergeSlice(&c.Inheritable, other.Inheritable) || changed
	changed = mergeSlice(&c.Bounding, other.Bounding) || changed
	changed = mergeSlice(&c.Ambient, other.Ambient) || changed
	return changed
}
