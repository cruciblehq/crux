package manifest

import "github.com/cruciblehq/crex"

// Current state format version.
const StateVersion = 0

// Represents the current state of a deployment.
//
// Records what resources have been deployed and their runtime identifiers.
// Used for incremental deployments and resource lifecycle management.
type State struct {
	Version    int        `codec:"version"`    // Version of the state format.
	Deployment Deployment `codec:"deployment"` // Metadata about the most recent deployment.
	Services   []Ref      `codec:"services"`   // Services that were deployed.
}

// Validates the state.
//
// The version must match [StateVersion]. The deployment timestamp must be set.
// Every service must have an ID and a ref.
func (s *State) Validate() error {
	if s.Version != StateVersion {
		return crex.Wrap(ErrInvalidState, ErrUnsupportedStateVersion)
	}

	if s.Deployment.DeployedAt.IsZero() {
		return crex.Wrap(ErrInvalidState, ErrMissingDeployedAt)
	}

	for i := range s.Services {
		if err := s.Services[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidState, err)
		}
		if s.Services[i].ID == "" {
			return crex.Wrap(ErrInvalidState, ErrMissingServiceID)
		}
	}

	return nil
}
