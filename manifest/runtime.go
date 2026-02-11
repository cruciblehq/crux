package manifest

// Holds configuration specific to runtime resources.
//
// Runtime resources define reusable base images for the Crucible ecosystem.
// They wrap external OCI images and apply additional setup (installing
// packages, copying configuration files, setting environment variables, etc.)
// to produce a base that service resources build on top of.
type Runtime struct {
	Recipe `yaml:",squash"`
}
