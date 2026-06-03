package manifest

import "github.com/cruciblehq/crux/crex"

// Current plan format version.
const PlanVersion = 0

// Represents a deployment plan.
//
// Specifies what resources will be deployed and the infrastructure required to
// run them. Generated during the planning phase by resolving references,
// compiling affordances, and determining routing.
//
// Infrastructure holds separate compute and network pools, each keyed by ID.
// Containers and Environments are lookup maps keyed by ID. Deployments is an
// ordered slice that binds them into deployable units.
type Plan struct {

	// Version of the plan format.
	Version int `codec:"version"`

	// Service lookup table. Key is the service ID, value is the reference target.
	Services map[string]string `codec:"services,omitempty"`

	// Infrastructure resource pools for compute and network allocations.
	Infrastructure Infrastructure `codec:"infrastructure"`

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
// The version must match [PlanVersion]. Infrastructure entries and environments
// are validated individually. Each deployment's field values are validated,
// then its service, container, compute, network, and environment references are
// checked against the plan's lookup tables. The gateway is validated last.
func (p *Plan) Validate() error {
	if p.Version != PlanVersion {
		return crex.Wrap(ErrInvalidPlan, ErrUnsupportedPlanVersion)
	}

	for _, target := range p.Services {
		if target == "" {
			return crex.Wrap(ErrInvalidPlan, ErrMissingRefTarget)
		}
	}

	if err := p.Infrastructure.Validate(); err != nil {
		return crex.Wrap(ErrInvalidPlan, err)
	}

	for _, e := range p.Environments {
		if err := e.Validate(); err != nil {
			return crex.Wrap(ErrInvalidPlan, err)
		}
	}

	for id, c := range p.Containers {
		if err := c.Validate(); err != nil {
			return crex.Wrapf(ErrInvalidPlan, "container %q: %w", id, err)
		}
	}

	for i := range p.Deployments {
		if err := p.validateDeployment(&p.Deployments[i]); err != nil {
			return err
		}
	}

	if err := p.Gateway.Validate(); err != nil {
		return crex.Wrap(ErrInvalidPlan, err)
	}

	return nil
}

// Validates a deployment's field values and checks that all of its binding
// references resolve to entries in the plan's lookup tables.
func (p *Plan) validateDeployment(d *Deployment) error {
	if err := d.Validate(); err != nil {
		return crex.Wrap(ErrInvalidPlan, err)
	}
	if _, ok := p.Services[d.Service]; !ok {
		return crex.Wrap(ErrInvalidPlan, crex.Wrapf(ErrUnresolvedDeploymentService, "%q", d.Service))
	}
	if _, ok := p.Containers[d.Container]; !ok {
		return crex.Wrap(ErrInvalidPlan, crex.Wrapf(ErrUnresolvedDeploymentContainer, "%q", d.Container))
	}
	if _, ok := p.Infrastructure.Computes[d.Compute]; !ok {
		return crex.Wrap(ErrInvalidPlan, crex.Wrapf(ErrUnresolvedDeploymentCompute, "%q", d.Compute))
	}
	if _, ok := p.Infrastructure.Networks[d.Network]; !ok {
		return crex.Wrap(ErrInvalidPlan, crex.Wrapf(ErrUnresolvedDeploymentNetwork, "%q", d.Network))
	}
	if d.Environment != "" {
		if _, ok := p.Environments[d.Environment]; !ok {
			return crex.Wrap(ErrInvalidPlan, crex.Wrapf(ErrUnresolvedDeploymentEnvironment, "%q", d.Environment))
		}
	}
	return nil
}
