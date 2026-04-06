package resource

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
)

// Verifies that a build directory contains the expected artifacts.
//
// Every resource type must have a valid manifest matching the expected type.
// When artifactFile is non-empty, verifies that the file also exists in the
// build directory.
func verify(buildDir string, resourceType manifest.ResourceType, artifactFile string) error {
	if _, err := verifyBuildDir(buildDir, resourceType); err != nil {
		return err
	}

	if artifactFile == "" {
		return nil
	}

	path := filepath.Join(buildDir, artifactFile)
	if _, err := os.Stat(path); err != nil {
		return crex.UserError("build output not found", fmt.Sprintf("%s/%s does not exist", buildDir, artifactFile)).
			Fallback("Run 'crux build' first to generate the build artifacts.").
			Cause(err).
			Err()
	}

	return nil
}

// Reads the manifest from a build directory and verifies its resource type.
//
// This should be called as the first step in [Builder.Verify] to ensure the
// build directory is of the right resource type before checking type-specific
// artifacts. Returns the manifest if its type matches the expected type.
func verifyBuildDir(buildDir string, expected manifest.ResourceType) (*manifest.Manifest, error) {
	manifestPath := filepath.Join(buildDir, manifest.ManifestFile)
	if _, err := os.Stat(manifestPath); err != nil {
		return nil, crex.UserError("manifest not found", "build/crucible.yaml does not exist").
			Fallback("Run 'crux build' first to generate the build artifacts.").
			Cause(err).
			Err()
	}

	m, err := ReadManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	if m.Resource.Type != expected {
		return nil, crex.UserError("resource type mismatch",
			fmt.Sprintf("expected %s but build directory contains %s", expected, m.Resource.Type)).
			Fallback("Ensure you are running the correct command for this resource type.").
			Err()
	}

	return m, nil
}
