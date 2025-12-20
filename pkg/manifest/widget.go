package manifest

const (

	// Default distribution directory for widget build output.
	WidgetDistDirectory = "dist"

	// Default distribution file for widget build output.
	WidgetDistMain = WidgetDistDirectory + "/" + "index.js"
)

// Holds configuration specific to widget resources.
//
// Widget resources are frontend components that can be embedded into apps.
// This structure defines configurations that are unique to widget resource
// manifests, such as build settings and requested affordances. It is used as
// the Config field in [Manifest] when the resource type is "widget".
type Widget struct {

	// Holds build-related configuration for widget resources.
	Build struct {
		Main string `field:"main"` // Build entry point.
	} `field:"build"`
}
