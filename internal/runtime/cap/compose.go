package cap

import (
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Merges the accumulated capability sets into the target [specs.LinuxCapabilities].
//
// Each accumulated set is unioned into the matching field of target, deduped
// against existing entries. A nil target is a no-op.
func (b *Builder) Compose(target *specs.LinuxCapabilities) error {
	if target == nil {
		return nil
	}

	addCaps(&target.Effective, b.caps.Effective)
	addCaps(&target.Permitted, b.caps.Permitted)
	addCaps(&target.Inheritable, b.caps.Inheritable)
	addCaps(&target.Bounding, b.caps.Bounding)
	addCaps(&target.Ambient, b.caps.Ambient)

	return nil
}
