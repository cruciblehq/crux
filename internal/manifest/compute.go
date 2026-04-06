package manifest

// Represents a compute resource in the deployment plan.
//
// Defines the compute instance to provision. The Config field contains
// provider-specific configuration based on the Provider value.
type Compute struct {

	// Stable identifier for this compute resource.
	ID string `codec:"id"`

	// Infrastructure provider (e.g. "aws", "local").
	Provider string `codec:"provider"`

	// Provider-specific configuration ([ComputeAWS] or [ComputeLocal]).
	Config any `codec:"config,omitempty"`
}
