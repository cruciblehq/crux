package resource

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/crex"
)

// Verifies that a build directory contains the expected artifacts.
//
// Every resource type requires a valid manifest matching t. Each type also
// requires its main artifact: runtimes and services an image, widgets a main
// bundle, and blueprints a plan.
func Verify(buildDir string, t manifest.ResourceType) error {
	switch t {
	case manifest.TypeRuntime:
		return verify(buildDir, t, filepath.Join(buildDir, manifest.ImageFile))
	case manifest.TypeService:
		return verify(buildDir, t, filepath.Join(buildDir, manifest.ImageFile))
	case manifest.TypeWidget:
		return verify(buildDir, t, filepath.Join(buildDir, manifest.WidgetMainFile))
	case manifest.TypeBlueprint:
		return verify(buildDir, t, filepath.Join(buildDir, manifest.PlanFile))
	default:
		return crex.Newf(ErrUnsupportedType, "resource type %q is not supported", t)
	}
}

// Verifies that a build directory contains the expected artifacts.
//
// Every resource type must have a valid manifest matching the expected type.
// When artifact is not empty, verifies that the file at that path also exists.
func verify(buildDir string, resourceType manifest.ResourceType, artifact string) error {
	if _, err := verifyBuildDir(buildDir, resourceType); err != nil {
		return err
	}

	if artifact == "" {
		return nil
	}

	if _, err := os.Stat(artifact); err != nil {
		return crex.Wrap(ErrBuildOutputNotFound, err)
	}

	return nil
}

// Reads the manifest from a build directory and verifies its resource type.
//
// This is the first step in [Verify], ensuring the build directory is of the
// right resource type before checking type-specific artifacts. Returns the
// manifest if its type matches the expected type.
func verifyBuildDir(buildDir string, expected manifest.ResourceType) (*manifest.Manifest, error) {
	m, err := ReadManifest(buildDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, crex.Wrap(ErrManifestNotFound, err)
	}
	if err != nil {
		return nil, err
	}

	if m.Resource.Type != expected {
		return nil, crex.Newf(ErrResourceTypeMismatch, "manifest declares resource type %q where %q is expected", m.Resource.Type, expected)
	}

	return m, nil
}
