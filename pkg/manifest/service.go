package manifest

// Holds configuration specific to service resources.
//
// Service resources are backend components that provide functionality to other
// systems by exposing an API. This structure defines configurations that are
// unique to service resources, such as container image references.
type Service struct {

	// Holds container image information for the service. Crux uses this to
	// locate the service's container image, which will later be deployed.
	Image struct {
		Ref string `key:"ref"` // Container image reference
	} `key:"image"`
}
