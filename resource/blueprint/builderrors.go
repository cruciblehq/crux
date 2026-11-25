package blueprint

import "github.com/cruciblehq/utils-go/crex"

// Wraps err in a build error that names the service id.
func errService(id string, err error) error {
	return crex.Wrapf(ErrBuildPlan, err, "service %s", id)
}

// Wraps err in a build error that names the service id and the directory
// holding its manifest.
func errServiceAt(id, dir string, err error) error {
	return crex.Wrapf(ErrBuildPlan, err, "service %s with manifest in %s", id, dir)
}
