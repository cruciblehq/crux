package resource

import (
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/paths"
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
		return crex.Wrap(ErrBuildOutputNotFound, err)
	}

	return nil
}

// Reads the manifest from a build directory and verifies its resource type.
//
// This should be called as the first step in [Handler.Verify] to ensure the
// build directory is of the right resource type before checking type-specific
// artifacts. Returns the manifest if its type matches the expected type.
func verifyBuildDir(buildDir string, expected manifest.ResourceType) (*manifest.Manifest, error) {
	manifestPath := paths.Manifest(buildDir)
	if _, err := os.Stat(manifestPath); err != nil {
		return nil, crex.Wrap(ErrManifestNotFound, err)
	}

	m, err := manifest.Read(manifestPath)
	if err != nil {
		return nil, err
	}

	if m.Resource.Type != expected {
		return nil, crex.Wrapf(ErrResourceTypeMismatch, "expected %s but got %s", expected, m.Resource.Type)
	}

	return m, nil
}
