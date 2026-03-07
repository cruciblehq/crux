package resource

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/cache"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/spec/archive"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/protocol"
	"github.com/cruciblehq/spec/reference"
)

// Common state and build logic for resource types that embed recipes.
type recipeBuilder struct {
	client   BuildClient // Daemon client for sending build requests.
	defaults Defaults    // Fallback registry and namespace for references.
	workdir  string      // Directory containing the manifest, root for resolving copy sources.
}

// Builds a recipe by resolving sources and delegating execution to cruxd.
//
// All stage sources are resolved to local file paths (pulling and extracting
// remote references as needed). The resolved recipe is sent to the daemon
// as a [protocol.BuildRequest]. The daemon handles container creation, step
// execution, and image export.
func (b *recipeBuilder) build(ctx context.Context, m manifest.Manifest, recipe *manifest.Recipe, output string, entrypoint []string) (*BuildResult, error) {
	if b.client == nil {
		return nil, crex.ProgrammingError("build failed", "daemon client is nil").
			Fallback("Ensure a compute node is running and pass a daemon client when constructing this builder.").
			Err()
	}

	resolved, cleanup, err := resolveAllSources(ctx, recipe, b.defaults)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	req := &protocol.BuildRequest{
		Recipe:     resolved,
		Resource:   m.Resource.Name,
		Output:     output,
		Root:       b.workdir,
		Entrypoint: entrypoint,
	}

	result, err := b.client.Build(ctx, req)
	if err != nil {
		return nil, crex.Wrap(ErrBuild, err)
	}

	return &BuildResult{Output: result.Output, Manifest: &m}, nil
}

// Resolves all stage sources in a recipe to forms the daemon can handle.
//
// Crucible runtime references (space-separated name and version constraint)
// are pulled, extracted, and rewritten to file paths. File and OCI sources
// are passed through unchanged — OCI references (single-token image names)
// are resolved by the daemon at build time. Returns a copy of the recipe
// with resolved sources, a cleanup function to remove temporary extraction
// directories, and any error encountered during resolution.
func resolveAllSources(ctx context.Context, recipe *manifest.Recipe, defaults Defaults) (*manifest.Recipe, func(), error) {
	resolved := *recipe
	resolved.Stages = make([]manifest.Stage, len(recipe.Stages))
	copy(resolved.Stages, recipe.Stages)

	var tempDirs []string
	cleanup := func() {
		for _, dir := range tempDirs {
			os.RemoveAll(dir)
		}
	}

	for i := range resolved.Stages {
		stage := &resolved.Stages[i]

		src, err := stage.ParseFrom()
		if err != nil {
			cleanup()
			return nil, nil, crex.Wrapf(ErrBuild, "stage %s: %w", stageLabel(stage.Name, i), err)
		}

		if src.Type == manifest.SourceRef {
			filePath, extractDir, err := resolveRefSource(ctx, src.Value, defaults)
			if err != nil {
				cleanup()
				return nil, nil, crex.Wrapf(ErrBuild, "stage %s: %w", stageLabel(stage.Name, i), err)
			}
			if extractDir != "" {
				tempDirs = append(tempDirs, extractDir)
			}
			stage.From = fmt.Sprintf("file %s", filePath)
		}
	}

	return &resolved, cleanup, nil
}

// Resolves a runtime reference to a local OCI image file path.
//
// Pulls the archive from the registry (with caching), extracts it to a
// temporary directory, and returns the path to the image file within the
// extracted archive. The caller must clean up the temporary directory.
func resolveRefSource(ctx context.Context, rawRef string, defaults Defaults) (imagePath, extractDir string, err error) {
	ref, err := reference.Parse(rawRef, string(manifest.TypeRuntime), defaults.ReferenceOptions())
	if err != nil {
		return "", "", err
	}

	result, err := Pull(ctx, ref)
	if err != nil {
		return "", "", err
	}

	localCache, err := cache.Open()
	if err != nil {
		return "", "", err
	}
	defer localCache.Close()

	archiveReader, err := localCache.OpenArchive(result.Namespace, result.Resource, result.Version)
	if err != nil {
		return "", "", err
	}
	defer archiveReader.Close()

	// Use the cache directory as the temp base so extracted archives stay
	// under the user's home directory. On macOS this is critical because the
	// home directory is the virtiofs mount shared with the build VM; the
	// system temp directory (/var/folders) is not mounted and would be
	// invisible to cruxd.
	tempBase := paths.CacheDir()
	if err := os.MkdirAll(tempBase, paths.DefaultDirMode); err != nil {
		return "", "", err
	}

	extractDir, err = os.MkdirTemp(tempBase, "crux-runtime-*")
	if err != nil {
		return "", "", err
	}

	if err := archive.ExtractFromReader(archiveReader, extractDir, archive.Zstd); err != nil {
		os.RemoveAll(extractDir)
		return "", "", err
	}

	return filepath.Join(extractDir, manifest.ImageFile), extractDir, nil
}

// Returns a label for a stage, preferring the name when available and falling
// back to the 1-based index.
func stageLabel(name string, index int) string {
	if name != "" {
		return fmt.Sprintf("%q", name)
	}
	return fmt.Sprintf("%d", index+1)
}
