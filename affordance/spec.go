package affordance

import (
	"github.com/cruciblehq/crux/affordance/fcap"
	"github.com/cruciblehq/crux/affordance/kernel"
	"github.com/cruciblehq/crux/affordance/mac"
	"github.com/cruciblehq/crux/affordance/net"
	"github.com/cruciblehq/crux/affordance/volume"
)

// Media type for a serialized compiled affordance spec.
//
// Used as the artifactType when the non-OCI portion of a spec is attached to a
// service image as an OCI referrer or stored as a typed content-store blob, so
// the runtime enforcement plugin can identify and resolve it.
const MediaType = "application/vnd.crucible.affordance.v0"

// Mutable runtime model accumulated by the affordance builder.
//
// Subsystems for caps, rlimits, seccomp, and cgroup mutate OCI directly; they
// are wired to fields of OCI at construction time. Fcap, MAC, Net, and Volume
// hold custom section models outside the OCI spec because OCI has no equivalent
// surface. Cgroup grants are projected into OCI.Linux.Resources: device entries
// go into Devices (the v2 BPF program is synthesised from them), every other
// knob goes into the Unified map in v2 kernel format. The kernel subsystem
// populates Kernel with the kernel requirements.
type Spec struct {
	OCI    *OCI         `codec:"oci,omitempty"`    // OCI runtime spec.
	Fcap   *fcap.Spec   `codec:"fcap,omitempty"`   // File capabilities per binary path.
	MAC    *mac.Spec    `codec:"mac,omitempty"`    // MAC (LSM) hook allow rules.
	Kernel *kernel.Spec `codec:"kernel,omitempty"` // Kernel requirements for the VM image.
	Net    *net.Spec    `codec:"net,omitempty"`    // Container network spec.
	Volume *volume.Spec `codec:"volume,omitempty"` // Persistent storage volume declarations.
}

// Returns a new unified Spec initialised to the strictest possible state.
//
// The OCI section carries a deny-all baseline that subsystems can only loosen.
// Custom non-OCI sections start in their zero-grant state.
func NewSpec() *Spec {
	return &Spec{
		OCI:    newOCIBaseline(),
		Fcap:   &fcap.Spec{Entries: make(map[string]*fcap.Capabilities)},
		MAC:    &mac.Spec{},
		Kernel: &kernel.Spec{},
		Net:    &net.Spec{},
		Volume: &volume.Spec{},
	}
}
