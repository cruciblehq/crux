package resource

import (
	"os"
	"path/filepath"

	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Verifies that a build directory contains the expected artifacts.
//
// Every resource type requires a valid manifest matching t. Types that emit a
// signature artifact also require that file: runtimes and services produce an
// image, widgets a main bundle, and blueprints a plan. Affordances have no
// additional artifact.
func Verify(buildDir string, t manifest.ResourceType) error {
	switch t {
	case manifest.TypeRuntime:
		return verify(buildDir, t, file.ImageFile)
	case manifest.TypeService:
		return verify(buildDir, t, file.ImageFile)
	case manifest.TypeWidget:
		return verify(buildDir, t, file.WidgetMainFile)
	case manifest.TypeAffordance:
		return verify(buildDir, t, "")
	case manifest.TypeBlueprint:
		return verify(buildDir, t, file.PlanFile)
	default:
		return crex.Newf(ErrUnsupportedType, "resource type %q is not supported", t)
	}
}

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
// This is the first step in [Verify], ensuring the build directory is of the
// right resource type before checking type-specific artifacts. Returns the
// manifest if its type matches the expected type.
func verifyBuildDir(buildDir string, expected manifest.ResourceType) (*manifest.Manifest, error) {
	manifestPath := file.Manifest(buildDir)
	if _, err := os.Stat(manifestPath); err != nil {
		return nil, crex.Wrap(ErrManifestNotFound, err)
	}

	m, err := manifest.Read(manifestPath)
	if err != nil {
		return nil, err
	}

	if m.Resource.Type != expected {
		return nil, crex.Newf(ErrResourceTypeMismatch, "expected %s but got %s", expected, m.Resource.Type)
	}

	return m, nil
}
