package manifest

// Local compute configuration.
//
// No additional configuration needed for local deployments.
type ComputeLocal struct{}

// Validates that the compute resource has an ID and provider.
func (c *Compute) Validate() error {
	if c.ID == "" {
		return ErrMissingComputeID
	}
	if c.Provider == "" {
		return ErrMissingProvider
	}
	return nil
}
