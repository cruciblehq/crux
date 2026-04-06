package manifest

// Defines a container in the deployment plan.
//
// A container assigns a service to a compute instance, optionally injecting
// an environment variable set. Grants holds the resolved grants produced
// during plan resolution.
type Container struct {
	Service     string  `codec:"service"`               // Service ID to run in this container.
	Compute     string  `codec:"compute"`               // Compute resource ID to run the container on.
	Environment string  `codec:"environment,omitempty"` // Environment set ID to inject (optional).
	Grants      []Grant `codec:"grants,omitempty"`      // Resolved grants for this container.
}

// Validates that the container references a service and compute resource.
func (c *Container) Validate() error {
	if c.Service == "" {
		return ErrMissingContainerService
	}
	if c.Compute == "" {
		return ErrMissingContainerCompute
	}
	return nil
}
