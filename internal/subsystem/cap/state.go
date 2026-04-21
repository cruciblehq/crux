package cap

import (
	"slices"

	"github.com/cruciblehq/crex"
)

// Accumulated capability state across all five kernel sets.
//
// Linux capabilities divide root privilege into individual units. Each field
// holds the names of capabilities granted to one of the five kernel sets:
// effective, permitted, inheritable, bounding, and ambient. The mode on each
// incoming rule determines which sets are populated. Capabilities are purely
// additive and idempotent. No conflicts are possible.
type State struct {
	Effective   []string `codec:"effective"`   // Effective capability set.
	Permitted   []string `codec:"permitted"`   // Permitted capability set.
	Inheritable []string `codec:"inheritable"` // Inheritable capability set.
	Bounding    []string `codec:"bounding"`    // Bounding capability set.
	Ambient     []string `codec:"ambient"`     // Ambient capability set.
}

// Returns a new capability state.
func NewState() *State {
	return &State{}
}

// Clones a capability state, returning nil if the state is nil or empty.
//
// This is used to avoid serialising empty capability sets in the affordance
// output. The State struct is serialised with the "omitempty" codec tag,
// which only includes non-empty fields in the output.
func cloneState(state *State) *State {
	if state == nil {
		return nil
	}
	if len(state.Effective) == 0 && len(state.Permitted) == 0 &&
		len(state.Inheritable) == 0 && len(state.Bounding) == 0 &&
		len(state.Ambient) == 0 {
		return nil
	}
	return &State{
		Effective:   slices.Clone(state.Effective),
		Permitted:   slices.Clone(state.Permitted),
		Inheritable: slices.Clone(state.Inheritable),
		Bounding:    slices.Clone(state.Bounding),
		Ambient:     slices.Clone(state.Ambient),
	}
}

// Merges a rule into the accumulated state.
//
// Uses the rule's mode to determine which sets are populated. Returns true
// if any capability was effectively added to any set.
func (s *State) Apply(r *Grant) (bool, error) {
	switch r.Mode {
	case ModeFull:
		return s.grantFull(r.Name), nil
	case ModeEffective:
		return s.grantEffective(r.Name), nil
	case ModeInheritable:
		return s.grantInheritable(r.Name), nil
	case ModePermitted:
		return s.grantPermitted(r.Name), nil
	case ModeBound:
		return s.grantBound(r.Name), nil
	default:
		return false, crex.Wrapf(ErrInvalidRule, "unknown capability mode %q", r.Mode)
	}
}

// Grants a capability to all five sets.
//
// The capability is effective immediately, survives exec, and auto-inherits
// to child processes. This is the broadest grant.
func (s *State) grantFull(name string) bool {
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
func (s *State) grantEffective(name string) bool {
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
func (s *State) grantInheritable(name string) bool {
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
func (s *State) grantPermitted(name string) bool {
	changed := appendUnique(&s.Permitted, name)
	changed = appendUnique(&s.Bounding, name) || changed
	return changed
}

// Grants a capability only in the bounding set.
//
// This acts as an exec ceiling: child processes may receive this capability
// (via file caps or ambient), but the current process cannot use it.
func (s *State) grantBound(name string) bool {
	return appendUnique(&s.Bounding, name)
}

// Incorporates all capabilities from another state.
//
// Each capability is added to the appropriate sets based on which sets it is
// present in the other state. Repeated capabilities are silently ignored.
func (s *State) Merge(other *State) {
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
