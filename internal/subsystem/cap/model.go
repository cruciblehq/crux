package cap

import (
	"slices"

	"github.com/cruciblehq/crex"
)

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

// Accumulated capability model across all five kernel sets.
//
// Linux capabilities divide root privilege into individual units. Each field
// holds the names of capabilities granted to one of the five kernel sets:
// effective, permitted, inheritable, bounding, and ambient. The mode on each
// incoming rule determines which sets are populated. Capabilities are purely
// additive and idempotent. No conflicts are possible.
type Model struct {
	Effective   []string `codec:"effective"`   // Effective capability set.
	Permitted   []string `codec:"permitted"`   // Permitted capability set.
	Inheritable []string `codec:"inheritable"` // Inheritable capability set.
	Bounding    []string `codec:"bounding"`    // Bounding capability set.
	Ambient     []string `codec:"ambient"`     // Ambient capability set.
}

// Returns a new capability model.
func NewModel() *Model {
	return &Model{}
}

// Clones a capability model, returning nil if the model is nil or empty.
//
// This is used to avoid serialising empty capability sets in the affordance
// output. The Model struct is serialised with the "omitempty" codec tag,
// which only includes non-empty fields in the output.
func cloneModel(m *Model) *Model {
	if m == nil {
		return nil
	}
	if len(m.Effective) == 0 && len(m.Permitted) == 0 &&
		len(m.Inheritable) == 0 && len(m.Bounding) == 0 &&
		len(m.Ambient) == 0 {
		return nil
	}
	return &Model{
		Effective:   slices.Clone(m.Effective),
		Permitted:   slices.Clone(m.Permitted),
		Inheritable: slices.Clone(m.Inheritable),
		Bounding:    slices.Clone(m.Bounding),
		Ambient:     slices.Clone(m.Ambient),
	}
}

// Grants a capability to all five sets.
//
// The capability is effective immediately, survives exec, and auto-inherits
// to child processes. This is the broadest grant.
func (s *Model) grantFull(name string) bool {
	changed := appendUnique(&s.Effective, name)
	changed = appendUnique(&s.Permitted, name) || changed
	changed = appendUnique(&s.Inheritable, name) || changed
	changed = appendUnique(&s.Bounding, name) || changed
	changed = appendUnique(&s.Ambient, name) || changed
	return changed
}

// Grants a capability to the effective, permitted, and bounding sets.
//
// The capability is effective immediately and survives exec (via bounding),
// but does not auto-inherit to child processes.
func (s *Model) grantEffective(name string) bool {
	changed := appendUnique(&s.Effective, name)
	changed = appendUnique(&s.Permitted, name) || changed
	changed = appendUnique(&s.Bounding, name) || changed
	return changed
}

// Grants a capability that auto-inherits across exec.
//
// The capability is not effective in the current process, but after execve
// the ambient set automatically raises it into the child's effective and
// permitted sets.
func (s *Model) grantInheritable(name string) bool {
	changed := appendUnique(&s.Permitted, name)
	changed = appendUnique(&s.Inheritable, name) || changed
	changed = appendUnique(&s.Ambient, name) || changed
	changed = appendUnique(&s.Bounding, name) || changed
	return changed
}

// Grants a capability to the permitted and bounding sets.
//
// The process may raise it into its effective set at will, and the bounding
// set allows it to persist across exec. Not effective by default, and does
// not auto-inherit.
func (s *Model) grantPermitted(name string) bool {
	changed := appendUnique(&s.Permitted, name)
	changed = appendUnique(&s.Bounding, name) || changed
	return changed
}

// Grants a capability only in the bounding set.
//
// This acts as an exec ceiling: child processes may receive this capability
// (via file caps or ambient), but the current process cannot use it.
func (s *Model) grantBound(name string) bool {
	return appendUnique(&s.Bounding, name)
}

// Incorporates all capabilities from another model.
//
// Each capability is added to the appropriate sets based on which sets it is
// present in the other model. Repeated capabilities are silently ignored.
func (s *Model) Merge(other *Model) {
	if other == nil {
		return
	}
	for _, name := range other.Effective {
		appendUnique(&s.Effective, name)
	}
	for _, name := range other.Permitted {
		appendUnique(&s.Permitted, name)
	}
	for _, name := range other.Inheritable {
		appendUnique(&s.Inheritable, name)
	}
	for _, name := range other.Bounding {
		appendUnique(&s.Bounding, name)
	}
	for _, name := range other.Ambient {
		appendUnique(&s.Ambient, name)
	}
}

// Appends s to the slice if not already present. Returns true if added.
func appendUnique(dst *[]string, s string) bool {
	if slices.Contains(*dst, s) {
		return false
	}
	*dst = append(*dst, s)
	return true
}
