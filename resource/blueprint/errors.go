package blueprint

import (
	"errors"

	"github.com/cruciblehq/crux/crex"
)

var ErrBuild = errors.New("blueprint build failed")

// Wraps err with a service-scoped build error.
func errService(id string, err error) error {
	return crex.Wrapf(ErrBuild, "service %s: %w", id, err)
}

// Wraps err with a service-runtime-scoped build error.
func errServiceRuntime(id string, err error) error {
	return crex.Wrapf(ErrBuild, "service %s: runtime: %w", id, err)
}
