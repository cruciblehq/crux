package blueprint

import "github.com/cruciblehq/utils-go/crex"

// Wraps err in a build error that names the service id.
func errService(id string, err error) error {
	return crex.Wrapf(ErrBuildPlan, err, "service %s", id)
}
