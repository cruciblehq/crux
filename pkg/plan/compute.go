package plan

import (
	"github.com/cruciblehq/crux/pkg/config"
	"github.com/cruciblehq/protocol/pkg/plan"
)

const (

	// Default compute ID (placeholder).
	DefaultComputeID = "main-compute"

	// Default compute instance type for all AWS deployments.
	DefaultAWSInstanceType = "t3.micro"
)

// Allocates compute resources for the deployment plan.
//
// Creates a single compute instance with provider-specific configuration. For
// AWS, includes instance type. For local, no additional config is needed.
func allocateCompute(p *plan.Plan, providerType config.ProviderType, providerName string) {
	compute := plan.Compute{
		ID:       DefaultComputeID,
		Provider: providerName,
		Config:   computeConfigForProvider(providerType),
	}
	p.Compute = append(p.Compute, compute)
}

// Returns provider-specific compute configuration.
func computeConfigForProvider(providerType config.ProviderType) any {
	switch providerType {
	case config.ProviderTypeAWS:
		return plan.ComputeAWS{
			InstanceType: DefaultAWSInstanceType,
		}
	case config.ProviderTypeLocal:
		return nil // No config needed for local
	default:
		return nil
	}
}
