package manifest

// Pool of compute and network resources allocated by the blueprint builder.
//
// The blueprint builder populates the pool based on the compute and network
// grants declared by services in the blueprint. Each grant contributes to the
// total resource requirements, which the builder uses to select and configure
// the compute and network resources needed to run the services. The pool is
// included in the plan so that the executor can provision the resources and
// apply the configurations when deploying the services.
type Infrastructure struct {

	// Allocated compute units.
	//
	// Each compute unit defines the virtual machine where a service runs. The
	// map associates compute unit IDs with their configurations, which can be
	// referenced by deployments. The executor provisions the compute units and
	// schedules services onto them according to the deployment specifications.
	Computes map[string]Compute `codec:"computes,omitempty"`

	// Allocated network configurations.
	//
	// Each network configuration defines the network settings for a service.
	// The map associates network configuration IDs with their settings, which
	// can be referenced by deployments. The executor applies the configurations
	// to the services' containers according to the deployment specifications.
	Networks map[string]Network `codec:"networks,omitempty"`
}

// Validates all compute and network entries in the pool.
func (i *Infrastructure) Validate() error {
	for _, c := range i.Computes {
		if err := c.Validate(); err != nil {
			return err
		}
	}
	for _, n := range i.Networks {
		if err := n.Validate(); err != nil {
			return err
		}
	}
	return nil
}
