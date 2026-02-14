package manifest

// Holds configuration specific to service resources.
//
// Service resources are backend components that provide functionality to other
// systems by exposing an API. They build on top of a base image defined by
// the embedded [Recipe], which specifies the source image and build steps.
type Service struct {
	Recipe `yaml:",squash"`

	// Command to run when the container starts.
	//
	// Sets the entrypoint on the output image produced by the recipe.
	Entrypoint []string `yaml:"entrypoint,omitempty"`
}
