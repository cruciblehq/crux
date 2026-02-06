package manifest

// Holds configuration specific to widget resources.
//
// Widget resources are frontend components that can be embedded into apps.
// This structure defines configurations that are unique to widget resource
// manifests, such as build settings and requested affordances. It is used as
// the Config field in [Manifest] when the resource type is "widget".
type Widget struct {
	Main string `yaml:"main"` // Build entry point.
}
