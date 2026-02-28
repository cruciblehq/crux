package resource

import (
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/spec/manifest"
)

// Reads and decodes the manifest at the given path.
func ReadManifest(manifestPath string) (*manifest.Manifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, crex.Wrap(ErrReadManifest, err)
	}
	m, err := manifest.Decode(data)
	if err != nil {
		return nil, crex.Wrap(ErrReadManifest, err)
	}
	return m, nil
}

// Serializes a manifest and writes it to a directory.
func WriteManifest(m *manifest.Manifest, dir string) error {
	data, err := manifest.Encode(m)
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(dir, manifest.ManifestFile)
	if err := os.WriteFile(manifestPath, data, paths.DefaultFileMode); err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	return nil
}
