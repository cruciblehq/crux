package manifest

import (
	"github.com/cruciblehq/crux/codec"
	"github.com/cruciblehq/crux/crex"
)

// Defines a service resource manifest.
//
// Service resources are backend components that provide functionality to other
// systems by exposing an API. They build on top of a base image defined by
// the embedded [Recipe], which specifies the source image and build steps.
type Service struct {
	Recipe

	// Declared parameters for this service.
	//
	// Lists configuration values the service accepts at runtime. Values are
	// bound through environment declarations.
	Schema *Schema `codec:"schema,omitempty"`

	// Command to run when the container starts.
	//
	// Sets the entrypoint on the output image produced by the recipe. The
	// entrypoint is required for service resources since they are expected
	// to run as containers.
	Entrypoint []string `codec:"entrypoint,omitempty"`
}

// Encodes the service configuration to a format-independent map.
//
// The schema, entrypoint, and recipe stages are encoded in the flat grant
// format so that grants serialize correctly.
func (s *Service) Encode(c *codec.Codec) (any, error) {
	m := make(map[string]any)
	stages, err := s.Recipe.encodeStages(c)
	if err != nil {
		return nil, err
	}
	if stages != nil {
		m["stages"] = stages
	}
	if s.Schema != nil {
		schema, err := c.ToMap(s.Schema)
		if err != nil {
			return nil, err
		}
		if len(schema) > 0 {
			m["schema"] = schema
		}
	}
	if len(s.Entrypoint) > 0 {
		m["entrypoint"] = s.Entrypoint
	}
	return m, nil
}

// Validates the service configuration.
func (s *Service) Validate() error {
	if len(s.Entrypoint) == 0 {
		return crex.Wrap(ErrInvalidService, ErrMissingEntrypoint)
	}

	if s.Schema != nil {
		if err := s.Schema.Validate(); err != nil {
			return crex.Wrap(ErrInvalidService, err)
		}
	}

	return s.Recipe.Validate()
}
