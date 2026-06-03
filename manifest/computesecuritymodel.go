package manifest

import (
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/security/vm"
)

// Compiled VM-level security model for a compute unit.
//
// Produced by the blueprint builder after bin-packing assigns services to
// compute units. The VM requirements from each service's affordance grants are
// unioned into a single policy that the provider must enforce at provisioning
// time. Each provider translates the policy into its native enforcement
// mechanism; no provider-specific fields appear here.
type ComputeSecurityModel struct {

	// VM security requirements derived from service grants.
	//
	// Union of the VM specs accumulated by all service affordance builders
	// assigned to this compute unit. Kernel features must be present in the
	// VM image; sysctls are applied at boot time; nftables rules are loaded
	// into the VM-level deny-all firewall before any container starts.
	VM vm.VM `codec:"vm"`
}

// Validates the security model.
func (p *ComputeSecurityModel) Validate() error {
	if err := p.VM.Validate(); err != nil {
		return crex.Wrap(ErrInvalidComputeSecurityModel, err)
	}
	return nil
}
