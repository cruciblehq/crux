package manifest

import "github.com/cruciblehq/crux/internal/crex"

// Current plan format version.
const PlanVersion = 0

// Represents a deployment plan.
//
// Specifies what resources will be deployed and the infrastructure configuration
// required to run them. Generated during the planning phase by resolving
// references, allocating infrastructure, and determining routing.
type Plan struct {
	Version      int           `json:"version"`                // Version of the plan format.
	Services     []Ref         `json:"services"`               // Services included in the deployment.
	Compute      []Compute     `json:"compute"`                // Compute resources to provision.
	Environments []Environment `json:"environments,omitempty"` // Environment variable sets for service configuration.
	Containers   []Container   `json:"containers"`             // Containers to deploy.
	Gateway      Gateway       `json:"gateway"`                // Gateway routing configuration.
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
