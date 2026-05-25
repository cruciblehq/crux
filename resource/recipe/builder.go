package recipe

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/paths"
	"github.com/cruciblehq/crux/source"
)

// Orchestrates a recipe build pipeline against a compute backend.
//
// Drives the stage pipeline: imports base images, compiles stage affordances
// into a security policy, and executes each step in order. The final image is
// exported as an OCI tar archive.
type Builder struct {
	src     source.Source // Registry access for pulling base images and resolving affordances.
	workdir string        // Manifest directory.
	backend Backend       // Performs the OCI image operations.
}

// Returns a new Builder.
//
// source provides registry access for pulling base images and resolving
// affordance references. workdir is the directory containing the manifest
// and is the root for resolving copy step sources. backend performs the
// OCI image operations.
func NewBuilder(src source.Source, workdir string, backend Backend) *Builder {
	return &Builder{
		src:     src,
		workdir: workdir,
		backend: backend,
	}
}

// Executes a recipe and writes the output image as an OCI tar archive.
//
// Each stage resolves its base image, then compiles its grants into a security
// policy and executes its steps in order. The image produced by the last stage
// is exported to [paths.BuildImage]. Returns the build directory on success.
func (b *Builder) Run(ctx context.Context, m manifest.Manifest, recipe *manifest.Recipe, output string, entrypoint []string) (string, error) {
	stageImages := make(map[string]string)
	var currentRef string

	for i := range recipe.Stages {
		var err error
		currentRef, err = b.runStage(ctx, i+1, &recipe.Stages[i], stageImages)
		if err != nil {
			return "", err
		}
	}

	currentRef, err := b.setEntrypoint(ctx, currentRef, entrypoint)
	if err != nil {
		return "", err
	}

	return b.exportImage(ctx, currentRef, output)
}

// Sets the image entrypoint.
//
// The OCI image spec only allows setting the entrypoint at the image level, but
// the recipe allows it to be set at the step level. This function applies the
// final entrypoint as an image config update after all stages and steps have
// been executed.
func (b *Builder) setEntrypoint(ctx context.Context, imageRef string, entrypoint []string) (string, error) {
	if len(entrypoint) == 0 {
		return imageRef, nil
	}

	updated, err := b.backend.Configure(ctx, imageRef, &compute.ConfigUpdate{
		SetEntrypoint: entrypoint,
	})
	if err != nil {
		return "", crex.Wrap(ErrBuild, err)
	}
	return updated, nil
}

// Exports the final image as an OCI tar archive.
//
// The image is exported to [paths.ImageFile] within the output directory.
// Returns the output directory path on success.
func (b *Builder) exportImage(ctx context.Context, imageRef string, output string) (string, error) {
	if err := os.MkdirAll(output, paths.DefaultDirMode); err != nil {
		return "", crex.Wrap(ErrFileSystemOperation, err)
	}

	image := filepath.Join(output, paths.ImageFile)
	f, err := os.OpenFile(image, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, paths.DefaultFileMode)
	if err != nil {
		return "", crex.Wrap(ErrFileSystemOperation, err)
	}
	defer f.Close()

	if err := b.backend.Export(ctx, imageRef, f); err != nil {
		return "", crex.Wrap(ErrBuild, err)
	}

	return output, nil
}
