package spec

import (
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/manifest"
	cspec "github.com/cruciblehq/crux/spec"
	afnet "github.com/cruciblehq/crux/resource/affordance/net"
)

// Mutable runtime model accumulated by the affordance builder.
//
// Subsystems for caps, rlimits, seccomp, and cgroup mutate OCI directly; they
// are wired to fields of OCI at construction time. Fcap and MAC hold custom
// section models outside the OCI spec because OCI has no equivalent surface.
// Cgroup grants are projected into OCI.Linux.Resources: device entries go into
// Devices (the v2 BPF program is synthesised from them), every other knob goes
// into the Unified map in v2 kernel format. The VM subsystem populates VM with
// the kernel features, sysctls, and nftables rules required by the workload.
type Spec struct {
	OCI  *specs.Spec   `json:"oci"`            // OCI runtime spec.
	Fcap *cspec.Fcap `json:"fcap,omitempty"` // File capabilities per binary path.
	MAC  *cspec.MAC  `json:"mac,omitempty"`  // MAC (LSM) hook allow rules.
	VM   *cspec.VM   `json:"vm,omitempty"`   // VM image builder requirements.
	Net  *afnet.Spec      `json:"net,omitempty"`  // Container network policy.
}

// Returns a new unified Spec initialised to the strictest possible state.
//
// The OCI section carries a deny-all baseline that subsystems can only loosen.
// Custom non-OCI sections start in their zero-grant state.
func NewSpec() *Spec {
	return &Spec{
		OCI:  newOCIBaseline(),
		Fcap: &cspec.Fcap{Entries: make(map[string]*cspec.FcapCapabilities)},
		MAC:  &cspec.MAC{},
		Net:  &afnet.Spec{},
	}
}

// Returns a manifest.Container from this Spec.
//
// Used at the end of a build session to produce the serializable plan artifact.
func (s *Spec) ToSpec() manifest.Container {
	ctr := manifest.Container{}
	if s.OCI != nil {
		ctr.OCI = *s.OCI
	}
	if s.Fcap != nil {
		ctr.Fcap = *s.Fcap
	}
	if s.MAC != nil {
		ctr.MAC = *s.MAC
	}
	if s.Net != nil {
		ctr.Network = cspec.NetworkPolicy{
			Ingress: s.Net.Ingress,
			Egress:  s.Net.Egress,
		}
	}
	return ctr
}
