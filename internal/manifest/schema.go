package manifest

import "github.com/cruciblehq/crex"

// Declares the parameters a resource accepts.
//
// Params lists the named parameters. Default optionally names one of them
// as the recipient of scalar values. When a caller passes a plain value
// instead of a named argument map, it is assigned to the default parameter.
// If Default is set, it must reference an existing param. Zero value means
// no parameters.
type Schema struct {

	// Name of the parameter that receives scalar values.
	//
	// When a caller passes a plain value instead of a named argument map,
	// the value is assigned to this parameter. Empty means no default.
	Default string `codec:"default,omitempty"`

	// Named parameters accepted by the resource.
	//
	// Each param must have a unique name. Zero value means no parameters.
	Params []Param `codec:"params,omitempty"`
}

// Validates the schema.
//
// All params must be valid, names must be unique, and if Default is set
// it must reference an existing param.
func (s *Schema) Validate() error {
	seen := make(map[string]bool, len(s.Params))

	for i := range s.Params {
		if err := s.Params[i].Validate(); err != nil {
			return crex.Wrapf(ErrInvalidParam, "param %d: %w", i+1, err)
		}

		if seen[s.Params[i].Name] {
			return crex.Wrapf(ErrInvalidParam, "param %q: %w", s.Params[i].Name, ErrDuplicateParamName)
		}
		seen[s.Params[i].Name] = true
	}

	if s.Default != "" && !seen[s.Default] {
		return crex.Wrapf(ErrInvalidParam, "default %q: %w", s.Default, ErrDefaultNotInSchema)
	}

	return nil
}
