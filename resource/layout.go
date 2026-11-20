package resource

import (
	"os"
	"path/filepath"

	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/file"
)

// Locations in a resource project.
const (
	buildDirName = "build"           // Build output directory.
	distDirName  = "dist"            // Distributable archive directory.
	packageFile  = "package.tar.zst" // Packaged resource archive.
)

// Returns the build output directory of the project in dir.
//
// Its contents depend on the resource type. Widgets produce a JavaScript
// bundle, runtimes and services an OCI image, and blueprints a plan.
func BuildDir(dir string) string {
	return filepath.Join(dir, buildDirName)
}

// Returns the path to the packaged archive of the project in dir.
func PackagePath(dir string) string {
	return filepath.Join(dir, distDirName, packageFile)
}

// Reads the manifest in dir.
//
// Returns an error that wraps [os.ErrNotExist] when dir has no manifest, and
// [manifest.ErrDecodeFailed] when the manifest is malformed.
func ReadManifest(dir string) (*manifest.Manifest, error) {
	f, err := os.Open(filepath.Join(dir, manifest.ManifestFile))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return manifest.Decode(f)
}

// Writes m as the manifest in dir, replacing any manifest already there.
//
// The file is replaced atomically, so a failed write leaves the old manifest.
func WriteManifest(dir string, m *manifest.Manifest) error {
	a, err := file.CreateAtomic(dir, file.DefaultFileMode)
	if err != nil {
		return err
	}
	defer a.Abort()
	if err := manifest.Encode(a, m); err != nil {
		return err
	}
	return a.Commit(filepath.Join(dir, manifest.ManifestFile))
}
