package manifest

// A service deployment.
//
// Associates a service with the environment it runs on, including the container
// sandbox, compute unit, and network configuration. Each field is a reference
// to an entry in the plan's lookup tables. The executor resolves the references
// and applies the corresponding configurations when deploying the service. The
// environment reference is optional since not all services declare parameters.
type Deployment struct {

	// ID of the service to run.
	//
	// The service must be declared in the blueprint and resolved during the
	// build, otherwise the plan is invalid. The service's manifest provides
	// the configuration and affordances for the deployment.
	Service string `codec:"service"`

	// ID of the container runtime spec.
	//
	// The container runtime spec defines the sandbox environment for the
	// service. The spec is compiled from the service manifest during the
	// build and stored in the plan's Containers map. The executor uses
	// the spec to set up the container environment for the service.
	Container string `codec:"container"`

	// ID of the environment set to inject.
	//
	// An environment provides concrete values for the parameters declared
	// by the service. An environment with this ID must be declared in the
	// blueprint. The executor injects the environment variables into the
	// service's container at runtime.
	Environment string `codec:"environment,omitempty"`

	// ID of the compute unit to run on.
	//
	// The compute unit defines the virtual machine where the service runs. A
	// compute unit with this ID must be declared in the plan's computes map.
	// The executor schedules the service onto the specified compute unit.
	Compute string `codec:"compute"`

	// ID of the network configuration to use.
	//
	// Defines the network settings for the service. A network configuration
	// with this ID must be declared in the plan's networks map. The executor
	// applies the network configuration to the service's container.
	Network string `codec:"network"`
}

// Validates a deployment.
//
// The fields are only validated for presence; the referenced entries in the
// plan's lookup tables are not validated here. The environment reference is
// optional since not all services declare parameters.
func (d *Deployment) Validate() error {
	if d.Service == "" {
		return ErrMissingDeploymentService
	}
	if d.Container == "" {
		return ErrMissingDeploymentContainer
	}
	if d.Compute == "" {
		return ErrMissingDeploymentCompute
	}
	if d.Network == "" {
		return ErrMissingDeploymentNetwork
	}
	return nil
}
