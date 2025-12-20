package manifest

// Holds configuration specific to service resources.
//
// Service resources are backend components that provide functionality to other
// systems by exposing an API. This structure defines configurations that are
// unique to service resources, such as container image references.
type Service struct {

	// Holds build-related configuration for service resources.
	Build struct {
		Main string            `field:"main"` // Build entry point (e.g., Dockerfile).
		Args map[string]string `field:"args"` // Build arguments (ARG in Dockerfile).
	} `field:"build"`
}
