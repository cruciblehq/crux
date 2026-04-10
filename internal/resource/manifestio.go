package resource

import (
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/codec"
	"github.com/cruciblehq/crux/internal/manifest"
)

// Reads and decodes the manifest inside the given directory.
//
// Looks for the standard manifest filename ([manifest.ManifestFile]) inside
// dir. Use [ReadManifest] when you already have the full path.
func ReadManifestIn(dir string) (*manifest.Manifest, error) {
	return ReadManifest(filepath.Join(dir, manifest.ManifestFile))
}

// Reads and decodes the manifest at the given path.
func ReadManifest(manifestPath string) (*manifest.Manifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, crex.Wrap(ErrReadManifest, err)
	}
	var m manifest.Manifest
	if err := codec.Unmarshal(data, &m, codec.YAML); err != nil {
		return nil, crex.Wrap(ErrReadManifest, err)
	}
	return &m, nil
}

// Serializes a manifest and writes it to a directory.
func WriteManifest(m *manifest.Manifest, dir string) error {
	data, err := codec.Encode(m, codec.YAML)
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(dir, manifest.ManifestFile)
	if err := os.WriteFile(manifestPath, data, paths.DefaultFileMode); err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	return nil
}
