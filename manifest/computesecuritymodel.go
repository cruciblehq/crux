package manifest

import (
	"github.com/cruciblehq/crux/affordance/kernel"
	"github.com/cruciblehq/crux/crex"
)

// Compiled VM-level security model for a compute unit.
//
// This model is produced by the blueprint builder after bin-packing assigns
// services to compute units. The kernel requirements from each service's
// affordance grants are unioned into a single model that the provider must
// enforce at provisioning time. Each provider translates the model into its
// native enforcement mechanism; no provider-specific fields appear here.
type ComputeSecurityModel struct {

	// Kernel requirements derived from service grants.
	//
	// Union of the kernel specs accumulated by all service affordance builders
	// assigned to this compute unit. Each requirement must be satisfied by the
	// VM image.
	Kernel kernel.Spec `codec:"kernel"`
}

// Validates the security model.
func (p *ComputeSecurityModel) Validate() error {
	if err := p.Kernel.Validate(); err != nil {
		return crex.Wrap(ErrInvalidComputeSecurityModel, err)
	}
	return nil
}
