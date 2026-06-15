package blueprint

import "github.com/cruciblehq/crux/crex"

// Wraps err with a service-scoped build error.
func errService(id string, err error) error {
	return crex.Wrapf(ErrBuildPlan, err, "service %s", id)
}

// Wraps err with a service-runtime-scoped build error.
func errServiceRuntime(id string, err error) error {
	return crex.Wrapf(ErrBuildPlan, err, "runtime for service %s", id)
}
