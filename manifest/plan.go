package manifest

import "github.com/cruciblehq/crux/crex"

// Current plan format version.
const PlanVersion = 0

// Represents a deployment plan.
//
// Specifies what resources will be deployed and the infrastructure configuration
// required to run them. Generated during the planning phase by resolving
// references, allocating infrastructure, and determining routing.
//
// Services, Compute, Containers, and Environments are lookup maps keyed by ID.
// Deployments is an ordered slice that binds them into deployable units.
type Plan struct {

	// Version of the plan format.
	Version int `codec:"version"`

	// Service lookup table. Key is the service ID, value is the reference target.
	Services map[string]string `codec:"services,omitempty"`

	// Compute resource lookup table. Key is the compute ID.
	Compute map[string]Compute `codec:"compute,omitempty"`

	// Container runtime spec lookup table. Key is the container ID.
	Containers map[string]Container `codec:"containers,omitempty"`

	// Environment variable set lookup table. Key is the environment ID.
	Environments map[string]Environment `codec:"environments,omitempty"`

	// Ordered deployment arrangement.
	Deployments []Deployment `codec:"deployments,omitempty"`

	// Gateway routing configuration.
	Gateway Gateway `codec:"gateway"`
}

// Validates the plan.
//
// The version must match [PlanVersion]. Every service target and compute entry
// is validated, followed by every deployment and the gateway.
func (p *Plan) Validate() error {
	if p.Version != PlanVersion {
		return crex.Wrap(ErrInvalidPlan, ErrUnsupportedPlanVersion)
	}

	for _, target := range p.Services {
		if target == "" {
			return crex.Wrap(ErrInvalidPlan, ErrMissingRefTarget)
		}
	}

	for _, c := range p.Compute {
		if err := c.Validate(); err != nil {
			return crex.Wrap(ErrInvalidPlan, err)
		}
	}

	for _, e := range p.Environments {
		if err := e.Validate(); err != nil {
			return crex.Wrap(ErrInvalidPlan, err)
		}
	}

	for i := range p.Deployments {
		if err := p.Deployments[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidPlan, err)
		}
	}

	if err := p.Gateway.Validate(); err != nil {
		return crex.Wrap(ErrInvalidPlan, err)
	}

	return nil
}
