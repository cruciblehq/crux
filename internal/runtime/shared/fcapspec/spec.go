package fcapspec

import "slices"

// Selects how file capabilities are granted on a binary.
//
// File capabilities are extended attributes on executables that the kernel
// evaluates during execve to compute the new process's capability sets.
// The mode determines whether capabilities become file-permitted (effective
// on exec) or file-inheritable (only effective if the caller also holds
// them in its inheritable set).
type Mode string

const (
	ModeEffective   Mode = "effective"   // File-permitted + effective bit. Caps are immediately effective after exec.
	ModeInheritable Mode = "inheritable" // File-inheritable. Caps only effective if caller holds them in inheritable set.
)

// Whether m is a recognised mode value.
func (m Mode) IsValid() bool {
	return m == ModeEffective || m == ModeInheritable
}

// Accumulated file capability spec, keyed by binary path.
//
// Each path is associated with a [Capabilities] value containing the kernel's
// file capability extended attribute fields: file-permitted, file-inheritable,
// and the effective bit. Grants for the same path are merged: capability
// lists are unified and the effective bit is OR'd.
type Spec struct {
	Entries map[string]*Capabilities `json:"entries"` // Per-binary file capabilities.
}

// File capabilities for a single binary path.
//
// The kernel uses these fields during execve to compute the new process's
// effective capability set. File-permitted capabilities are granted to the
// new process; file-inheritable capabilities are granted only if the exec
// caller also holds them in its own inheritable set.
type Capabilities struct {
	Permitted   []string `json:"permitted"`   // Capabilities to grant as file-permitted.
	Inheritable []string `json:"inheritable"` // Capabilities to grant as file-inheritable.
	Effective   bool     `json:"effective"`   // If true, all granted file-permitted capabilities become immediately effective.
}

// Returns a new Spec with an empty Entries map.
func NewSpec() *Spec {
	return &Spec{Entries: make(map[string]*Capabilities)}
}

// Folds other into the receiver.
//
// For each path, capability lists are unified and the effective bit is OR'd.
// A nil other is a no-op.
func (s *Spec) MergeSpec(other *Spec) {
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
func (c *Capabilities) GrantEffective(caps []string) bool {
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
func (c *Capabilities) GrantInheritable(caps []string) bool {
	return mergeSlice(&c.Inheritable, caps)
}

// Appends elements from src that are not already in dst.
//
// Returns true if any element was added.
func mergeSlice(dst *[]string, src []string) bool {
	changed := false
	for _, s := range src {
		if !slices.Contains(*dst, s) {
			*dst = append(*dst, s)
			changed = true
		}
	}
	return changed
}
