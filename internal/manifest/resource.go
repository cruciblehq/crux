package manifest

import "github.com/cruciblehq/crex"

// Holds common metadata about the resource.
//
// Includes the resource type, qualified name, version, and an optional
// description. The type field determines how the rest of the manifest
// is interpreted.
type Resource struct {

	// The type of the resource.
	//
	// Determines how the rest of the manifest is interpreted and how
	// Crucible manages the resource.
	Type ResourceType `codec:"type"`

	// The qualified resource name.
	//
	// Identifies the resource, including its namespace, using the format
	// "namespace/name" (e.g. "cruciblehq/my-api"). The registry is not
	// part of the name; it is resolved from configuration.
	Name string `codec:"name"`

	// Description of the resource.
	//
	// Documents what the resource does or provides. Optional.
	Description string `codec:"description,omitempty"`

	// The version of the resource.
	//
	// This is a semantic version string that indicates the version of the
	// resource being defined. This field is required.
	Version string `codec:"version"`
}

// Validates the resource metadata.
//
// The type must be a known [ResourceType]. Name and version are required.
func (r *Resource) Validate() error {
	if _, err := ParseResourceType(string(r.Type)); err != nil {
		return crex.Wrap(ErrInvalidResource, err)
	}

	if r.Name == "" {
		return crex.Wrap(ErrInvalidResource, ErrMissingName)
	}

	if r.Version == "" {
		return crex.Wrap(ErrInvalidResource, ErrMissingVersion)
	}

	return nil
}
