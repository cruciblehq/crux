package build

import (
	"log/slog"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/manifest"
)

// Defines the signature for a function that builds a specific resource type.
//
// It accepts the configuration struct as [any] and returns an error indicating
// the result of the build.
type buildCallback func(any) error

// Generic helper that adapts a typed build function into a generic [buildCallback]
// while ensuring type safety.
//
// If the provided configuration is not of the expected type, it returns a user
// error report indicating the type mismatch.
func assertManifestType[T any](fn func(T) error) buildCallback {
	return func(config any) error {

		// Type assertion
		if typedConfig, ok := config.(T); ok {
			return fn(typedConfig)
		}

		// Type mismatch
		return crex.ProgrammingError("build failed", "an internal configuration type mismatch occurred").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}
}

// Builds the resource specified in the given manifest.
//
// It resolves the appropriate build function based on the resource type defined
// in the manifest, and invokes that function to perform the build. It returns
// a [crex.ReportBuilder] indicating the success or failure of the build.
func BuildResource(options *manifest.Manifest) error {

	// Supported resource types and their build functions
	resourceTypes := map[string]buildCallback{
		"widget":  assertManifestType(BuildWidget),
		"service": assertManifestType(BuildService),
	}

	// Resolve build function from resource type
	buildFunc, ok := resourceTypes[options.Resource.Type]

	if !ok {
		return crex.UserErrorf("build failed", "invalid resource type '%s'", options.Resource.Type).
			Fallback("Change your manifest to use a supported resource type.").
			Err()
	}

	slog.Debug("building resource", "type", options.Resource.Type)

	// Build
	return buildFunc(options.Config)
}
