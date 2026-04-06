package resource

import "github.com/cruciblehq/crux/internal/manifest"

// Holds the output of a successful [Builder.Build] call.
type BuildResult struct {
	Output   string             // Directory where the build artifacts were written.
	Manifest *manifest.Manifest // The fully resolved manifest used for the build.
}
