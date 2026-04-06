package manifest

import "github.com/cruciblehq/crex"

// Holds configuration specific to service resources.
//
// Service resources are backend components that provide functionality to other
// systems by exposing an API. They build on top of a base image defined by
// the embedded [Recipe], which specifies the source image and build steps.
type Service struct {
	Recipe `codec:",squash"`

	// Declared parameters for this service.
	//
	// Lists configuration values the service accepts at runtime. Values are
	// bound through environment declarations.
	Schema Schema `codec:"schema,omitempty"`

	// Command to run when the container starts.
	//
	// Sets the entrypoint on the output image produced by the recipe.
	Entrypoint []string `codec:"entrypoint,omitempty"`
}

// Validates the service configuration.
func (s *Service) Validate() error {
	if len(s.Entrypoint) == 0 {
		return crex.Wrap(ErrInvalidService, ErrMissingEntrypoint)
	}

	if err := s.Schema.Validate(); err != nil {
		return crex.Wrap(ErrInvalidService, err)
	}

	return s.Recipe.Validate()
}
