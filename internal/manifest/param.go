package manifest

import "github.com/cruciblehq/crex"

// A single named parameter.
type Param struct {
	Name     string `codec:"name"`               // Argument key name.
	Required bool   `codec:"required,omitempty"` // Whether the caller must supply this argument.
}

// Validates the param.
func (p *Param) Validate() error {
	if p.Name == "" {
		return crex.Wrap(ErrInvalidParam, ErrMissingParamName)
	}
	return nil
}
