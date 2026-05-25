package manifest

import "github.com/cruciblehq/crux/crex"

// Common identity envelope shared by every resource manifest.
//
// Crux reads Type first when decoding a manifest, using it to determine which
// concrete struct the rest of the document maps to. Keeping these fields in a
// shared type means the toolstack can inspect any manifest for its type and
// metadata without needing to know about the specific resource types in advance.
type Resource struct {

	// Discriminator that controls how the rest of the manifest is decoded.
	//
	// Must be set before any other field is read. The namespace portion of the
	// type string determines which registry is consulted; the name portion
	// selects the resource kind within that registry.
	Type ResourceType `codec:"type"`

	// Identifier for the resource.
	//
	// Uses the format "namespace/name" (e.g. "cruciblehq/my-api"). The
	// registry host is intentionally absent so the same manifest can be
	// published to multiple registries without modification.
	Name string `codec:"name"`

	// Human-readable summary of the resource's purpose.
	//
	// Not interpreted by crux itself. Present for documentation and registry
	// engines that surface resource metadata.
	Description string `codec:"description,omitempty"`

	// Semantic version declared by this resource.
	//
	// Set by the resource author at publish time and immutable thereafter.
	// The registry files the artifact under this version so consumers can
	// reference an exact, stable artifact.
	Version string `codec:"version"`
}

// Validates the resource metadata.
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
