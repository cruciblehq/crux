package manifest

// Represents a compute resource in the deployment plan.
//
// Defines the compute instance to provision. The Config field contains
// provider-specific configuration based on the Provider value.
type Compute struct {
	ID       string `json:"id"`               // Stable identifier for this compute resource.
	Provider string `json:"provider"`         // Infrastructure provider (e.g. "aws", "local").
	Config   any    `json:"config,omitempty"` // Provider-specific configuration ([ComputeAWS] or [ComputeLocal]).
}
