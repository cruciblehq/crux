package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cruciblehq/crux/pkg/cache"
	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/pull"
	"github.com/cruciblehq/protocol/pkg/archive"
	"github.com/cruciblehq/protocol/pkg/manifest"
	"github.com/cruciblehq/protocol/pkg/oci"
	"github.com/cruciblehq/protocol/pkg/resource"
)

// Builds multi-platform OCI images for runtimes and services.
type ImageBuilder struct {
	registry   string          // Hub registry URL for pulling base runtimes.
	base       string          // Reference to base runtime (empty for scratch builds).
	files      []manifest.File // Files to include in the image.
	entrypoint []string        // Entrypoint command (nil for runtimes).
	output     string          // Output directory path.
}

// Creates a new instance of [ImageBuilder].
func NewImageBuilder(registry, base string, files []manifest.File, entrypoint []string, output string) *ImageBuilder {
	return &ImageBuilder{
		registry:   registry,
		base:       base,
		files:      files,
		entrypoint: entrypoint,
		output:     output,
	}
}

// Builds a multi-platform OCI image.
//
// This is the common workflow for building both runtimes and services. If a
// base reference is set, fetches the base runtime first. Then groups files by
// platform, validates that all required platforms are present (only for scratch
// builds), and builds the multi-platform image.
func (ib *ImageBuilder) Build(ctx context.Context) (*Result, error) {
	var baseImagePath string
	if ib.base != "" {
		var cleanup func()
		var err error
		baseImagePath, cleanup, err = ib.fetchBaseImage(ctx)
		if err != nil {
			return nil, err
		}
		defer cleanup()
	}

	platformFiles, sharedFiles, err := groupFilesByPlatform(ib.files)
	if err != nil {
		return nil, err
	}

	// Validate platforms only when building from scratch
	if baseImagePath == "" {
		if err := validatePlatforms(platformFiles); err != nil {
			return nil, err
		}
	}

	outputPath := filepath.Join(ib.output, archive.ImageFile)
	if err := buildResourceImage(baseImagePath, platformFiles, sharedFiles, ib.entrypoint, outputPath); err != nil {
		return nil, crex.UserError("failed to build image", err.Error()).
			Fallback("Check that all source files exist and the output directory is writable.").
			Cause(err).
			Err()
	}

	return &Result{Output: ib.output}, nil
}

// Groups files by platform and validates they exist on disk.
func groupFilesByPlatform(files []manifest.File) (map[string][]manifest.File, []manifest.File, error) {
	for _, file := range files {
		if err := checkFileExists(file.Src); err != nil {
			return nil, nil, err
		}
	}
	platformFiles, sharedFiles := manifest.GroupFilesByPlatform(files)
	return platformFiles, sharedFiles, nil
}

// Checks that all required platforms have at least one file.
//
// Resources built from scratch must provide platform-specific files for each
// required platform (linux/amd64, linux/arm64).
func validatePlatforms(platformFiles map[string][]manifest.File) error {
	var missing []string
	for _, platform := range oci.RequiredPlatforms() {
		if _, ok := platformFiles[platform]; !ok {
			missing = append(missing, platform)
		}
	}
	if len(missing) > 0 {
		return crex.UserError("missing required platforms", fmt.Sprintf("no platform-specific files for %s", strings.Join(missing, ", "))).
			Fallback(fmt.Sprintf("Provide files with platform set for all required platforms (%v).", oci.RequiredPlatforms())).
			Err()
	}
	return nil
}

// Checks that a file exists on disk.
func checkFileExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return crex.UserError("file not found", fmt.Sprintf("%s does not exist", path)).
				Fallback("Ensure all files referenced in the manifest exist.").
				Err()
		}
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	return nil
}

// Fetches a base image from the registry.
//
// Downloads the resource archive (or uses a cached copy), extracts it to a
// temporary directory, and returns the path to the OCI image tarball inside.
// The caller must invoke the returned cleanup function to remove the temporary
// directory when done.
func (ib *ImageBuilder) fetchBaseImage(ctx context.Context) (string, func(), error) {
	result, err := pull.Pull(ctx, pull.Options{
		Registry:  ib.registry,
		Reference: ib.base,
		Type:      resource.TypeRuntime,
	})
	if err != nil {
		return "", nil, crex.UserError("failed to fetch base runtime", err.Error()).
			Fallback("Ensure the reference is valid and the resource is published.").
			Cause(err).
			Err()
	}

	extractDir, cleanup, err := extractArchive(ctx, result)
	if err != nil {
		return "", nil, err
	}

	imagePath := filepath.Join(extractDir, resource.DistDirectory, archive.ImageFile)
	if _, err := os.Stat(imagePath); err != nil {
		cleanup()
		return "", nil, crex.UserError("base image not found", "dist/image.tar missing from archive").
			Fallback("The base resource may not be built correctly.").
			Err()
	}

	return imagePath, cleanup, nil
}

// Extracts a resource archive to a temporary directory.
//
// Returns the extraction directory path and a cleanup function that removes
// the temporary directory.
func extractArchive(ctx context.Context, result *pull.Result) (string, func(), error) {
	localCache, err := cache.Open(ctx, nil)
	if err != nil {
		return "", nil, crex.Wrap(ErrFileSystemOperation, err)
	}
	defer localCache.Close()

	archiveReader, err := localCache.OpenArchive(ctx, result.Namespace, result.Resource, result.Version)
	if err != nil {
		return "", nil, crex.UserError("failed to read archive", err.Error()).
			Fallback("The resource may be corrupted. Try 'crux pull' to re-download.").
			Cause(err).
			Err()
	}
	defer archiveReader.Close()

	tmpDir, err := os.MkdirTemp("", "crux-base-*")
	if err != nil {
		return "", nil, crex.Wrap(ErrFileSystemOperation, err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := archive.ExtractFromReader(archiveReader, extractDir); err != nil {
		os.RemoveAll(tmpDir)
		return "", nil, crex.UserError("failed to extract archive", err.Error()).
			Fallback("The archive may be corrupted.").
			Cause(err).
			Err()
	}

	cleanup := func() { os.RemoveAll(tmpDir) }
	return extractDir, cleanup, nil
}

// Builds a multi-platform OCI image for a Crucible resource.
//
// If basePath is empty, builds from scratch using platforms derived from the
// platformFiles map keys. If basePath is provided, extends the base image
// using its platforms. Files are added to each platform image, and an optional
// entrypoint is set if provided.
func buildResourceImage(
	basePath string,
	platformFiles map[string][]manifest.File,
	sharedFiles []manifest.File,
	entrypoint []string,
	outputPath string,
) error {
	builder := oci.NewMultiPlatformBuilder()

	if basePath == "" {
		if err := buildFromScratch(builder, platformFiles, sharedFiles, entrypoint); err != nil {
			return err
		}
	} else {
		if err := buildFromBase(builder, basePath, platformFiles, sharedFiles, entrypoint); err != nil {
			return err
		}
	}

	if err := builder.SaveTarball(outputPath); err != nil {
		return crex.Wrap(ErrLayoutWrite, err)
	}
	return nil
}

// Builds all platforms from scratch using the platform files map.
func buildFromScratch(
	builder *oci.MultiPlatformBuilder,
	platformFiles map[string][]manifest.File,
	sharedFiles []manifest.File,
	entrypoint []string,
) error {
	for platform, files := range platformFiles {
		if err := buildPlatformFromScratch(builder, platform, files, sharedFiles, entrypoint); err != nil {
			return err
		}
	}
	return nil
}

// Builds all platforms by extending a base image.
func buildFromBase(
	builder *oci.MultiPlatformBuilder,
	basePath string,
	platformFiles map[string][]manifest.File,
	sharedFiles []manifest.File,
	entrypoint []string,
) error {
	baseIdx, err := oci.ReadIndex(basePath)
	if err != nil {
		return crex.Wrap(ErrInvalidImage, err)
	}
	defer baseIdx.Close()

	platforms, err := baseIdx.Platforms()
	if err != nil {
		return crex.Wrap(ErrInvalidImage, err)
	}

	for platform := range platforms {
		files := platformFiles[platform] // may be nil, that's ok
		if err := buildPlatformFromBase(builder, baseIdx, platform, files, sharedFiles, entrypoint); err != nil {
			return err
		}
	}
	return nil
}

// Builds a single platform image from scratch.
func buildPlatformFromScratch(
	builder *oci.MultiPlatformBuilder,
	platform string,
	platformFiles, sharedFiles []manifest.File,
	entrypoint []string,
) error {
	osName, arch, err := oci.ParsePlatform(platform)
	if err != nil {
		return err
	}

	platformBuilder := builder.ForPlatform(osName, arch)

	if err := addFiles(platformBuilder, platformFiles, sharedFiles); err != nil {
		return err
	}

	if len(entrypoint) > 0 {
		platformBuilder.SetEntrypoint(entrypoint...)
	}

	return nil
}

// Builds a single platform image by extending a base image.
func buildPlatformFromBase(
	builder *oci.MultiPlatformBuilder,
	baseIdx *oci.Index,
	platform string,
	platformFiles, sharedFiles []manifest.File,
	entrypoint []string,
) error {
	osName, arch, err := oci.ParsePlatform(platform)
	if err != nil {
		return err
	}

	baseImage, err := baseIdx.LoadImage(osName, arch)
	if err != nil {
		return crex.Wrap(ErrInvalidImage, err)
	}

	platformBuilder, err := oci.NewBuilderFrom(baseImage.Underlying(), osName, arch)
	if err != nil {
		return err
	}

	if err := addFiles(platformBuilder, platformFiles, sharedFiles); err != nil {
		return err
	}

	if len(entrypoint) > 0 {
		platformBuilder.SetEntrypoint(entrypoint...)
	}

	img, err := platformBuilder.Image()
	if err != nil {
		return crex.Wrap(ErrImageBuild, err)
	}
	builder.AddImage(osName, arch, img)

	return nil
}

// Adds platform-specific and shared files to an image builder.
func addFiles(builder *oci.Builder, platformFiles, sharedFiles []manifest.File) error {
	for _, file := range platformFiles {
		if err := builder.AddMapping(file.Src, file.Dest); err != nil {
			return crex.Wrap(ErrLayerCreate, err)
		}
	}

	for _, file := range sharedFiles {
		if err := builder.AddMapping(file.Src, file.Dest); err != nil {
			return crex.Wrap(ErrLayerCreate, err)
		}
	}

	return nil
}
