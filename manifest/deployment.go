package manifest

// Records a service deployment arrangement.
//
// In a plan, captures which service artifact runs inside which container
// sandbox, which environment variable set is injected at startup, and which
// compute resource hosts the unit. All IDs reference entries in the parent
// Plan's lookup slices.
//
// In state, also records when the deployment was applied via DeployedAt.
type Deployment struct {
	Service     string `codec:"service"`               // ID of the service to run.
	Container   string `codec:"container"`             // ID of the container runtime spec.
	Environment string `codec:"environment,omitempty"` // ID of the environment set to inject.
	Compute     string `codec:"compute"`               // ID of the compute resource to run on.
}

// Validates that the deployment references a service, a container, and a compute resource.
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
	return nil
}
