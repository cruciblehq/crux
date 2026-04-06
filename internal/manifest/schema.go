package manifest

import "github.com/cruciblehq/crex"

// Declares the parameters a resource accepts.
//
// Params lists the named parameters. Default optionally names one of them
// as the recipient of scalar values (when a caller passes a plain value
// instead of a named argument map, it is assigned to the default parameter).
type Schema struct {
	Default string  `codec:"default,omitempty"` // Name of the param that receives scalar values.
	Params  []Param `codec:"params,omitempty"`  // Named parameters.
}

// Validates the schema.
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
