package fcap

import (
	"slices"

	"github.com/cruciblehq/crex"
)

// Accumulated file capability state, keyed by binary path.
//
// Each path is associated with a Capabilities value containing the kernel's
// file capability extended attribute fields: file-permitted, file-inheritable,
// and the effective bit. Grants for the same path are merged: capability lists
// are unified and the effective bit is OR'd.
type State struct {
	Entries map[string]*Capabilities `codec:"entries"` // Per-binary file capabilities.
}

// File capabilities for a single binary path.
//
// The kernel uses these fields during execve to compute the new process's
// effective capability set. File-permitted capabilities are granted to the
// new process; file-inheritable capabilities are granted only if the exec
// caller also holds them in its own inheritable set.
type Capabilities struct {
	Permitted   []string `codec:"permitted"`   // Capabilities to grant as file-permitted.
	Inheritable []string `codec:"inheritable"` // Capabilities to grant as file-inheritable.
	Effective   bool     `codec:"effective"`   // If true, all granted file-permitted capabilities become immediately effective.
}

// Returns a new State with an empty Entries map.
func NewState() *State {
	return &State{Entries: make(map[string]*Capabilities)}
}

// Returns a deep copy of state, or nil when state is nil or has no entries.
//
// The returned copy duplicates the Entries map and each Capabilities value.
// Capability slices are copied into new backing arrays to preserve isolation.
func cloneState(state *State) *State {
	if state == nil || len(state.Entries) == 0 {
		return nil
	}
	copy := NewState()
	for path, entry := range state.Entries {
		copy.Entries[path] = &Capabilities{
			Permitted:   append([]string(nil), entry.Permitted...),
			Inheritable: append([]string(nil), entry.Inheritable...),
			Effective:   entry.Effective,
		}
	}
	return copy
}

// Merges a rule into the accumulated state.
//
// Dispatches on the rule's mode to populate the appropriate fields of the
// Capabilities for the given path. Returns true if any capability was added or
// the effective bit was changed.
func (s *State) Apply(r *Grant) (bool, error) {
	if s.Entries == nil {
		s.Entries = make(map[string]*Capabilities)
	}
	e, ok := s.Entries[r.Path]
	if !ok {
		e = &Capabilities{}
	}
	switch r.Mode {
	case ModeEffective:
		s.Entries[r.Path] = e
		return e.grantEffective(r.Caps), nil
	case ModeInheritable:
		s.Entries[r.Path] = e
		return e.grantInheritable(r.Caps), nil
	default:
		return false, crex.Wrapf(ErrInvalidRule, "unknown fcap mode %q", r.Mode)
	}
}

// Incorporates all entries from another state.
//
// Merges all entries from the input state into the accumulated state. For
// each path, if an entry already exists, capability lists are unified and
// the effective bit is OR'd. If no entry exists, a new one is created with
// the same values as the input entry.
func (s *State) Merge(other *State) {
	if other == nil {
		return
	}
	if s.Entries == nil {
		s.Entries = make(map[string]*Capabilities)
	}
	for path, e := range other.Entries {
		existing, ok := s.Entries[path]
		if !ok {
			existing = &Capabilities{}
			s.Entries[path] = existing
		}
		mergeSlice(&existing.Permitted, e.Permitted)
		mergeSlice(&existing.Inheritable, e.Inheritable)
		if e.Effective {
			existing.Effective = true
		}
	}
}

// Grants file-permitted capabilities and sets the effective bit.
//
// After execve the capabilities are immediately effective in the new process.
func (c *Capabilities) grantEffective(caps []string) bool {
	changed := mergeSlice(&c.Permitted, caps)
	if !c.Effective {
		c.Effective = true
		changed = true
	}
	return changed
}

// Grants file-inheritable capabilities.
//
// The capabilities only take effect if the calling process also holds
// them in its inheritable set.
func (c *Capabilities) grantInheritable(caps []string) bool {
	return mergeSlice(&c.Inheritable, caps)
}

// Appends elements from src that are not already in dst.
//
// Returns true if any element was added.
func mergeSlice(dst *[]string, src []string) bool {
	changed := false
	for _, s := range src {
		found := slices.Contains(*dst, s)
		if !found {
			*dst = append(*dst, s)
			changed = true
		}
	}
	return changed
}
