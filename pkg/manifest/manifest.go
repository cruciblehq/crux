package manifest

// Holds metainformation about a resource.
//
// This structure both specifies metadata about the resource and influences the
// validation of the rest of the manifest structure. This structure is required
// and must always be the first field in the manifest.
type Manifest struct {

	// The manifest version.
	//
	// This is required and must be the first declaration in the manifest. This
	// value dictates how the rest of the manifest is interpreted. Currently,
	// the only supported version is 0.
	Version int `field:"version"`

	// Holds common metadata about the resource.
	//
	// This data structure includes all metadata that is shared across resource
	// types, including the resource type itself. This is required and must be
	// the second field in the manifest, after [Manifest.Version].
	Resource Resource `field:"resource"`

	// Lists affordances requested by the resource.
	//
	// Affordances are integrations or capabilities that the resource can request
	// from Crucible. They enable the resource to interact with other services or
	// utilize specific features provided by Crucible. This structure consists of
	// a list of key/value mappings where they key is an arbitrary label and the
	// value is the affordance type. Affordance configurations are passed to the
	// resource under the associated label.
	Affordances []map[string]string `field:"affordances,omitempty"`

	// Holds type-specific configuration, depending on the resource type.
	//
	// This field is polymorphic and its concrete type depends on the value of
	// [Manifest.Resource.Type]. For example, if the resource type is "widget",
	// this field will be of type [Widget]. If the resource type is "service",
	// this field will be of type [Service]. This field is required and must be
	// the last field in the manifest.
	Config any `field:"-"`
}

// Holds common metadata about the resource.
//
// This structure both specifies metadata about the resource and influences the
// rest of the manifest structure. This structure is required and must always be
// the second field in the manifest, after [Manifest.Version].
type Resource struct {

	// The type of the resource.
	//
	// This field determines how the rest of the manifest is interpreted, as
	// well as the behavior of Crucible when managing the resource.
	Type string `field:"type"`

	// The version of the resource.
	//
	// This is a semantic version string that indicates the version of the
	// resource being defined. This field is required.
	Version string `field:"version"`
}
