package build

import (
	"context"

	"github.com/cruciblehq/crux/manifest"
)

// Builder interface for different resource types.
//
// Each resource type (e.g., widget, service) should have its own implementation
// of this interface to handle the specific build process for that resource.
type Builder interface {

	// Builds the provided resource, performing the build process specific to
	// the resource type.
	//
	// The type declared on the manifest must match the builder's resource type.
	// If it does not, the build is aborted and an error is returned.
	Build(ctx context.Context, m manifest.Manifest, output string) (*Result, error)
}
