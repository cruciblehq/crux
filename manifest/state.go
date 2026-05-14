package manifest

import "github.com/cruciblehq/crux/crex"

// Current state format version.
const StateVersion = 0

// Represents the current state of a deployment.
//
// Records what resources have been deployed and their runtime identifiers.
// Used for incremental deployments and resource lifecycle management.
type State struct {

	// Version of the state format.
	Version int `codec:"version"`

	// Deployments that were applied.
	Deployments []Deployment `codec:"deployments"`
}

// Validates the state.
//
// The version must match [StateVersion]. The deployment timestamp must be set.
// Every service must have an ID and a ref.
func (s *State) Validate() error {
	if s.Version != StateVersion {
		return crex.Wrap(ErrInvalidState, ErrUnsupportedStateVersion)
	}

	for i := range s.Deployments {
		if err := s.Deployments[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidState, err)
		}
	}

	return nil
}
