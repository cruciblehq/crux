package manifest

import (
	"github.com/cruciblehq/crux/crex"
)

// A named parameter.
//
// Parameters define the arguments a resource accepts. Each parameter has
// a unique name within its schema. A parameter with a non-nil Default is
// optional; the default value is used when the caller omits it. A parameter
// with no default (nil) is required and must be supplied by the caller.
type Param struct {

	// Parameter name.
	//
	// Used as the key in the argument map passed by callers. Must be
	// non-empty and unique within the parent [Schema].
	Name string `codec:"name"`

	// Fallback value used when the caller does not supply this argument.
	//
	// When nil, the parameter is required and omitting it is a validation
	// error. When non-nil, the parameter is optional and the default value
	// is substituted.
	Default any `codec:"default,omitempty"`
}

// Validates the param.
//
// Name must be non-empty and match [a-z][a-z0-9_]*. Default, when set,
// must be a scalar (string, int, float64, or bool).
func (p *Param) Validate() error {
	if p.Name == "" {
		return crex.Wrap(ErrInvalidParam, ErrMissingParamName)
	}
	if !isValidName(p.Name) {
		return crex.Tag(crex.Newf(ErrInvalidParam, "param %q has an invalid name", p.Name), ErrInvalidParamName)
	}
	if p.Default != nil {
		switch p.Default.(type) {
		case string, int, float64, bool:
			// valid scalar types
		default:
			return crex.Tag(crex.Newf(ErrInvalidParam, "param %q has an invalid default", p.Name), ErrInvalidParamDefault)
		}
	}
	return nil
}
