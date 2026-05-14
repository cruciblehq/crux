package manifest

import (
	"github.com/cruciblehq/crux/codec"
	"github.com/cruciblehq/crux/crex"
)

// Runtime resources define reusable base images.
//
// They wrap external OCI images and apply additional setup (installing
// packages, copying configuration files, setting environment variables,
// etc.) to produce a base that service resources build on top of. The
// main difference between runtimes and services is that runtimes do not
// have an entrypoint since they are not expected to run as containers.
type Runtime struct {
	Recipe

	// Declared parameters for this runtime.
	//
	// Lists build-time configuration values the runtime accepts. Values are
	// bound through environment declarations.
	Schema *Schema `codec:"schema,omitempty"`
}

// Encodes the runtime configuration to a format-independent map.
//
// The schema and recipe stages are encoded in the flat grant format so that
// grants serialize correctly.
func (r *Runtime) Encode() (any, error) {
	m := make(map[string]any)
	stages, err := r.Recipe.encodeStages()
	if err != nil {
		return nil, err
	}
	if stages != nil {
		m["stages"] = stages
	}
	if r.Schema != nil {
		sm, err := codec.ToMap(r.Schema)
		if err != nil {
			return nil, err
		}
		if len(sm) > 0 {
			m["schema"] = sm
		}
	}
	return m, nil
}

// Validates the runtime configuration.
func (r *Runtime) Validate() error {
	if r.Schema != nil {
		if err := r.Schema.Validate(); err != nil {
			return crex.Wrap(ErrInvalidRecipe, err)
		}
	}

	return r.Recipe.Validate()
}
