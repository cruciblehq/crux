package plan

import (
	"github.com/cruciblehq/protocol/pkg/plan"
)

const (

	// Default compute ID (placeholder).
	DefaultComputeID = "main-compute"
)

// Allocates compute resources for the deployment plan.
//
// Creates a single compute instance with default configuration that all services
// will share. The instance uses t3.micro type and is provisioned on the specified
// provider.
func allocateCompute(p *plan.Plan, provider string) {
	computeID := DefaultComputeID
	compute := plan.Compute{
		ID:           computeID,
		Provider:     provider,
		InstanceType: DefaultInstanceType,
	}
	p.Compute = append(p.Compute, compute)
}
