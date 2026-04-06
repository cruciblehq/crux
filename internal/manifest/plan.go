package manifest

import "github.com/cruciblehq/crex"

// Current plan format version.
const PlanVersion = 0

// Represents a deployment plan.
//
// Specifies what resources will be deployed and the infrastructure configuration
// required to run them. Generated during the planning phase by resolving
// references, allocating infrastructure, and determining routing.
type Plan struct {
	Version      int           `codec:"version"`                // Version of the plan format.
	Services     []Ref         `codec:"services"`               // Services included in the deployment.
	Compute      []Compute     `codec:"compute"`                // Compute resources to provision.
	Environments []Environment `codec:"environments,omitempty"` // Environment variable sets for service configuration.
	Containers   []Container   `codec:"containers"`             // Containers to deploy.
	Gateway      Gateway       `codec:"gateway"`                // Gateway routing configuration.
}

// Validates the plan.
//
// The version must match [PlanVersion]. Every service must have an ID and ref,
// every compute must have an ID and provider, every binding must reference
// a service and compute, and every route must have a pattern and service.
func (p *Plan) Validate() error {
	if p.Version != PlanVersion {
		return crex.Wrap(ErrInvalidPlan, ErrUnsupportedPlanVersion)
	}

	for i := range p.Services {
		if err := p.Services[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidPlan, err)
		}
	}

	for i := range p.Compute {
		if err := p.Compute[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidPlan, err)
		}
	}

	for i := range p.Containers {
		if err := p.Containers[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidPlan, err)
		}
	}

	if err := p.Gateway.Validate(); err != nil {
		return crex.Wrap(ErrInvalidPlan, err)
	}

	return nil
}
