package manifest

import "github.com/cruciblehq/crex"

// Holds configuration specific to runtime resources.
//
// Runtime resources define reusable base images for the Crucible ecosystem.
// They wrap external OCI images and apply additional setup (installing
// packages, copying configuration files, setting environment variables, etc.)
// to produce a base that service resources build on top of.
type Runtime struct {
	Recipe `codec:",squash"`

	// Declared parameters for this runtime.
	//
	// Lists build-time configuration values the runtime accepts. Values are
	// bound through environment declarations.
	Schema Schema `codec:"schema,omitempty"`
}

// Validates the runtime configuration.
func (r *Runtime) Validate() error {
	if err := r.Schema.Validate(); err != nil {
		return crex.Wrap(ErrInvalidRecipe, err)
	}

	return r.Recipe.Validate()
}
