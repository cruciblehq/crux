package manifest

// Local machine configuration.
//
// The identity of a compute unit backed by a directly reachable host. Unlike
// cloud providers, no provisioning API is involved — the executor connects to
// the host as-is and schedules services onto it.
type ComputeLocal struct {

	// Hostname or IP address of the target machine.
	Host string `codec:"host,omitempty"`
}

// Validates the local compute configuration.
func (c *ComputeLocal) Validate() error {
	if c.Host == "" {
		return ErrMissingComputeHost
	}
	return nil
}
