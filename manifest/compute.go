package manifest

// Represents a compute resource in the deployment plan.
//
// Defines the compute instance to provision. The Config field contains
// provider-specific configuration based on the Provider value.
type Compute struct {
	Provider string `codec:"provider"`         // Infrastructure provider (e.g. "aws", "local").
	Config   any    `codec:"config,omitempty"` // Provider-specific configuration ([ComputeAWS] or [ComputeLocal]).
}

// Validates that the compute entry specifies a known provider.
func (c *Compute) Validate() error {
	if c.Provider == "" {
		return ErrMissingProvider
	}
	if _, err := ParseProviderType(c.Provider); err != nil {
		return err
	}
	return nil
}
