package manifest

import "github.com/cruciblehq/crux/crex"

// Named string arguments.
//
// Each key must be a valid name as defined by [isValidName]. Values are always
// plain strings; type coercion and interpolation may happen later, during the
// build stage.
type Args map[string]string

// Validates all keys in the argument map.
func (a Args) Validate() error {
	for k := range a {
		if !isValidName(k) {
			return crex.Wrapf(ErrInvalidArgKey, "arg key %q", k)
		}
	}
	return nil
}
