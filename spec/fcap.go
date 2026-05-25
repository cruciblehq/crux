package spec

import "slices"

// Accumulated file capability spec, keyed by binary path.
//
// Multiple affordances targeting the same binary path are merged into a single
// entry so the resulting map holds the minimal set of grants needed by the
// workload.
type Fcap struct {

	// Per-binary file capabilities.
	//
	// Keyed by absolute path in the container filesystem.
	Entries map[string]*FcapCapabilities `codec:"entries"`
}

// File capabilities for a single binary path.
//
// The kernel uses these fields during execve to compute the new process's
// effective capability set.
type FcapCapabilities struct {

	// Capabilities to grant as file-permitted.
	Permitted []string `codec:"permitted"`

	// Capabilities to grant as file-inheritable.
	Inheritable []string `codec:"inheritable"`

	// Immediately activates the file-permitted capabilities after execve.
	//
	// When true, the kernel sets the file effective bit so the new process does
	// not need to raise the capabilities itself.
	Effective bool `codec:"effective"`
}

// Selects how file capabilities are granted on a binary.
//
// Effective mode also sets the file effective bit, which causes the granted
// capabilities to be active immediately after execve without the new process
// having to raise them itself. Inheritable mode requires the calling process
// to already hold the capabilities in its inheritable set.
type FcapMode string

const (

	// File-permitted plus effective bit.
	//
	// Capabilities are immediately active after execve without the process
	// needing to raise them.
	FcapModeEffective FcapMode = "effective"

	// File-inheritable.
	//
	// Capabilities take effect only if the caller already holds them in its
	// inheritable set.
	FcapModeInheritable FcapMode = "inheritable"
)

// Whether m is a recognised mode value.
func (m FcapMode) IsValid() bool {
	return m == FcapModeEffective || m == FcapModeInheritable
}

// Grants file-permitted capabilities and sets the effective bit.
//
// After execve the capabilities are immediately effective in the new process.
// Returns true if any state was changed.
func (c *FcapCapabilities) GrantEffective(caps []string) bool {
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
func (c *FcapCapabilities) GrantInheritable(caps []string) bool {
	return mergeStringSlice(&c.Inheritable, caps)
}

// Appends elements from src that are not already in dst.
//
// Returns true if any element was added.
func mergeStringSlice(dst *[]string, src []string) bool {
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
