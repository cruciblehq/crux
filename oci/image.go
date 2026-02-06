package oci

import (
	"fmt"
	"os"

	"github.com/cruciblehq/crux/kit/crex"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
)

// Default permission bits for created directories.
const DirPermissions = 0755

// Wraps a v1.ImageIndex with optional cleanup of temporary resources.
//
// When reading from tarballs, temporary files are created. Callers must call
// Close() when done to release these resources. Use Underlying() to access
// the raw v1.ImageIndex for registry operations.
type Index struct {
	idx     v1.ImageIndex
	cleanup func()
}

// Returns the underlying v1.ImageIndex for use with go-containerregistry.
//
// Use this when you need to pass the index to external libraries like
// the remote package for registry operations.
func (i *Index) Underlying() v1.ImageIndex {
	return i.idx
}

// Returns the index manifest containing all platform descriptors.
func (i *Index) IndexManifest() (*v1.IndexManifest, error) {
	return i.idx.IndexManifest()
}

// Wraps a v1.Image for consistent API.
//
// Use Underlying() to access the raw v1.Image for registry operations.
type Image struct {
	img v1.Image
}

// Returns the underlying v1.Image for use with go-containerregistry.
//
// Use this when you need to pass the image to external libraries like
// the remote package for registry operations.
func (i *Image) Underlying() v1.Image {
	return i.img
}

// Returns the image configuration.
func (i *Image) ConfigFile() (*v1.ConfigFile, error) {
	return i.img.ConfigFile()
}

// Reads an OCI image index from a tarball or layout directory.
//
// Supports both OCI layout tarballs (tar archives of OCI layout directories)
// and plain OCI layout directories. Callers must call Close() on the returned
// Index to release any temporary resources.
func ReadIndex(imagePath string) (*Index, error) {
	info, err := os.Stat(imagePath)
	if err != nil {
		return nil, crex.Wrap(ErrInvalidImage, err)
	}

	if info.IsDir() {
		p, err := layout.FromPath(imagePath)
		if err != nil {
			return nil, crex.Wrap(ErrInvalidImage, err)
		}
		idx, err := p.ImageIndex()
		if err != nil {
			return nil, crex.Wrap(ErrInvalidImage, err)
		}
		return &Index{idx: idx}, nil
	}

	return readIndexFromTarball(imagePath)
}

// Releases any temporary resources associated with the index.
//
// Safe to call multiple times. Returns nil always.
func (i *Index) Close() error {
	if i.cleanup != nil {
		i.cleanup()
		i.cleanup = nil
	}
	return nil
}

// Extracts valid platform identifiers from the index.
//
// Excludes attestation manifests and descriptors with unknown os/architecture.
// Returns a map where keys are platform strings in "os/arch" format.
func (i *Index) Platforms() (map[string]bool, error) {
	manifest, err := i.idx.IndexManifest()
	if err != nil {
		return nil, crex.Wrap(ErrInvalidImage, err)
	}

	platforms := make(map[string]bool)
	for _, desc := range manifest.Manifests {
		if desc.Platform == nil {
			continue
		}
		// Skip attestation manifests with unknown os/arch
		if desc.Platform.OS == "unknown" || desc.Platform.Architecture == "unknown" {
			continue
		}
		key := fmt.Sprintf("%s/%s", desc.Platform.OS, desc.Platform.Architecture)
		platforms[key] = true
	}
	return platforms, nil
}

// Validates that the index contains all required platforms.
//
// Returns ErrInsufficientPlatforms if the image supports only one platform,
// none, or is missing any of the required platforms.
func (i *Index) ValidateMultiPlatform() error {
	platforms, err := i.Platforms()
	if err != nil {
		return err
	}

	if len(platforms) == 0 {
		return ErrInsufficientPlatforms
	}

	var missing []string
	for _, required := range RequiredPlatforms() {
		if !platforms[required] {
			missing = append(missing, required)
		}
	}

	if len(missing) > 0 {
		return ErrInsufficientPlatforms
	}

	return nil
}

// Loads an image for a specific platform from the index.
//
// Returns the image matching the specified os/arch combination, or an error
// if the platform is not found.
func (i *Index) LoadImage(osName, arch string) (*Image, error) {
	manifest, err := i.idx.IndexManifest()
	if err != nil {
		return nil, crex.Wrap(ErrInvalidImage, err)
	}

	for _, desc := range manifest.Manifests {
		if desc.Platform == nil {
			continue
		}
		if desc.Platform.OS == osName && desc.Platform.Architecture == arch {
			img, err := i.idx.Image(desc.Digest)
			if err != nil {
				return nil, crex.Wrap(ErrInvalidImage, err)
			}
			return &Image{img: img}, nil
		}
	}

	return nil, crex.Wrap(ErrInvalidImage, fmt.Errorf("platform %s/%s not found", osName, arch))
}

// Saves the index to disk as an OCI layout directory.
//
// Creates the target directory if it doesn't exist.
func (i *Index) SaveLayout(path string) error {
	if err := os.MkdirAll(path, DirPermissions); err != nil {
		return crex.Wrap(ErrLayoutWrite, err)
	}

	if _, err := layout.Write(path, i.idx); err != nil {
		return crex.Wrap(ErrLayoutWrite, err)
	}
	return nil
}

// Reads an image index from an OCI layout tarball.
//
// Extracts the tarball to a temporary directory and reads it as an OCI layout.
// The returned Index includes a cleanup function that removes the temp directory.
func readIndexFromTarball(tarballPath string) (*Index, error) {
	f, err := os.Open(tarballPath)
	if err != nil {
		return nil, crex.Wrap(ErrInvalidImage, err)
	}
	defer f.Close()

	tmpDir, err := os.MkdirTemp("", "oci-tarball-*")
	if err != nil {
		return nil, crex.Wrap(ErrInvalidImage, err)
	}

	if err := extractTar(f, tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}

	p, err := layout.FromPath(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, crex.Wrap(ErrInvalidImage, err)
	}

	idx, err := p.ImageIndex()
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, crex.Wrap(ErrInvalidImage, err)
	}

	return &Index{
		idx:     idx,
		cleanup: func() { os.RemoveAll(tmpDir) },
	}, nil
}
