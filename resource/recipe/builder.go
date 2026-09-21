package recipe

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/crux/resource/oci"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Author label recorded in the OCI history of the final exported image.
const commitAuthorExport = "export"

// Orchestrates a recipe build pipeline against a compute backend.
//
// Drives the stage pipeline: imports base images and executes each step in
// order. The final image is exported as an OCI tar archive.
type Builder struct {
	src     hub.Source      // Registry access for pulling base images.
	workdir string          // Manifest directory.
	client  *compute.Client // Live client connection to the container runtime on the build host.
}

// Returns a new Builder.
//
// source provides registry access for pulling base images. workdir is the
// directory containing the manifest and is the root for resolving copy step
// sources. client is the open connection to the container runtime on the host.
func NewBuilder(src hub.Source, workdir string, client *compute.Client) *Builder {
	return &Builder{
		src:     src,
		workdir: workdir,
		client:  client,
	}
}

// Executes a recipe and writes the output image as an OCI tar archive.
//
// Each stage resolves its base image and executes its steps in order. The
// image produced by the last stage is exported to [file.ImageFile]. When set,
// entrypoint becomes the image entrypoint. Returns the build directory.
func (b *Builder) Build(ctx context.Context, recipe *manifest.Recipe, entrypoint []string, output string) (string, error) {
	stageImages := make(map[string]string)
	var currentCtr *compute.Container

	for i := range recipe.Stages {
		if currentCtr != nil {
			currentCtr.Destroy(ctx)
		}
		ctr, err := b.runStage(ctx, i+1, &recipe.Stages[i], stageImages)
		if err != nil {
			return "", err
		}
		currentCtr = ctr
	}
	defer currentCtr.Destroy(ctx)

	if err := b.setEntrypoint(ctx, currentCtr, entrypoint); err != nil {
		return "", err
	}

	return b.exportImage(ctx, currentCtr, output)
}

// Sets the image entrypoint on the container before it is committed.
//
// The OCI image spec only allows setting the entrypoint at the image level,
// but the recipe allows it to be set at the step level. This function applies
// the final entrypoint as an image config update after all stages and steps
// have been executed.
func (b *Builder) setEntrypoint(ctx context.Context, ctr *compute.Container, entrypoint []string) error {
	if len(entrypoint) == 0 {
		return nil
	}
	cfg, err := ctr.Inspect(ctx)
	if err != nil {
		return crex.Wrap(ErrBuild, err)
	}
	cfg.Entrypoint = entrypoint
	ctr.Configure(cfg)
	return nil
}

// Exports the final image as an OCI tar archive.
//
// The image is exported to [file.ImageFile] within the output directory.
// Returns the output directory path on success.
func (b *Builder) exportImage(ctx context.Context, ctr *compute.Container, output string) (string, error) {
	if err := os.MkdirAll(output, file.DefaultDirMode); err != nil {
		return "", crex.Wrap(oci.ErrFileSystemOperation, err)
	}

	image := filepath.Join(output, file.ImageFile)
	f, err := os.OpenFile(image, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.DefaultFileMode)
	if err != nil {
		return "", crex.Wrap(oci.ErrFileSystemOperation, err)
	}
	defer f.Close()

	ref, err := ctr.Commit(ctx, commitAuthorExport)
	if err != nil {
		return "", crex.Wrap(ErrBuild, err)
	}
	if err := b.client.Export(ctx, ref, f); err != nil {
		return "", crex.Wrap(ErrBuild, err)
	}

	return output, nil
}
