package shared

import (
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/internal/runtime/shared/fcapspec"
	"github.com/cruciblehq/crux/internal/runtime/shared/macspec"
)

// Runtime model.
//
// Subsystems for caps, rlimits, seccomp, and cgroup mutate fields directly
// under OCI; they are wired at construction time. Fcap and MAC define custom,
// non-OCI section models held outside the OCI spec because OCI has no
// equivalent surface for them.
//
// Cgroup grants are projected into OCI.Linux.Resources: device entries are
// written to the typed Devices array (the v2 BPF program is synthesised from
// it), every other knob is written to the Unified map verbatim in v2 kernel
// format. The cgroup subsystem keeps a private typed model for conflict
// detection only; it does not appear here.
type Spec struct {
	OCI  *specs.Spec    `json:"oci"`            // Owning OCI runtime spec.
	Fcap *fcapspec.Spec `json:"fcap,omitempty"` // File capabilities per binary path.
	MAC  *macspec.Spec  `json:"mac,omitempty"`  // MAC (LSM) hook allow rules.
}

// Returns a new unified Spec initialised to the strictest possible state.
//
// The OCI section carries a deny-all baseline that subsystems can only
// loosen. Custom non-OCI sections start in their zero-grant state.
func NewSpec() *Spec {
	return &Spec{
		OCI:  newOCIBaseline(),
		Fcap: fcapspec.NewSpec(),
		MAC:  macspec.NewSpec(),
	}
}
