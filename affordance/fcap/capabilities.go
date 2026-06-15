package fcap

import (
	"slices"

	"github.com/cruciblehq/crux/crex"
)

// File capabilities for a single binary path.
//
// The kernel uses these fields during execve to compute the new process's
// effective capability set. This set is smaller than Linux cap sets, which
// also include bounding and inheritable sets.
type Capabilities struct {
	Permitted   []string `codec:"permitted"`   // Capabilities to grant as file-permitted.
	Inheritable []string `codec:"inheritable"` // Capabilities to grant as file-inheritable.
	Effective   bool     `codec:"effective"`   // Immediately activates the file-permitted capabilities after execve.
}

// Grants file-permitted capabilities and sets the effective bit.
//
// After execve the capabilities are immediately effective in the new process.
// Returns true if any state was changed.
func (c *Capabilities) GrantEffective(caps []string) bool {
	changed := mergeStringSlice(&c.Permitted, caps)
	if !c.Effective {
		c.Effective = true
		changed = true
	}
	return changed
}

// Grants file-inheritable capabilities.
//
// The capabilities only take effect if the calling process also holds
// them in its inheritable set. Returns true if any state was changed.
func (c *Capabilities) GrantInheritable(caps []string) bool {
	return mergeStringSlice(&c.Inheritable, caps)
}

// Validates the capabilities entry.
//
// At least one of Permitted or Inheritable must be non-empty.
// If Effective is true, Permitted must also be non-empty.
func (c *Capabilities) Validate() error {
	if len(c.Permitted) == 0 && len(c.Inheritable) == 0 {
		return crex.Wrapf(ErrInvalidCapabilities, "at least one of permitted or inheritable must be non-empty")
	}
	if c.Effective && len(c.Permitted) == 0 {
		return crex.Wrapf(ErrInvalidCapabilities, "effective requires non-empty permitted")
	}
	return nil
}

// Appends elements from src that are not already in dst.
//
// Returns true if any element was added.
func mergeStringSlice(dst *[]string, src []string) bool {
	changed := false
	for _, s := range src {
		if !slices.Contains(*dst, s) {
			*dst = append(*dst, s)
			changed = true
		}
	}
	return changed
}
