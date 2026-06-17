package recipe

import (
	"context"
	"os"
	"path/filepath"

	aff "github.com/cruciblehq/crux/affordance"
	"github.com/cruciblehq/crux/codec"
	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/files"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/registry"
	"github.com/cruciblehq/crux/resource/oci"
)

// Author label recorded in the OCI history of the final exported image.
const commitAuthorExport = "export"

// Orchestrates a recipe build pipeline against a compute backend.
//
// Drives the stage pipeline: imports base images, compiles stage affordances
// into a security spec, and executes each step in order. The final image is
// exported as an OCI tar archive.
type Builder struct {
	src     registry.Source // Registry access for pulling base images and resolving affordances.
	workdir string          // Manifest directory.
	client  *compute.Client // Live client connection to the container runtime on the build host.
}

// Returns a new Builder.
//
// source provides registry access for pulling base images and resolving
// affordance references. workdir is the directory containing the manifest
// and is the root for resolving copy step sources. client is the open
// connection to the container runtime on the build host.
func NewBuilder(src registry.Source, workdir string, client *compute.Client) *Builder {
	return &Builder{
		src:     src,
		workdir: workdir,
		client:  client,
	}
}

// Executes a recipe and writes the output image as an OCI tar archive.
//
// Each stage resolves its base image, then compiles its grants into a security
// spec and executes its steps in order. The image produced by the last stage
// is exported to [files.BuildImage]. entrypoint, when set, becomes the image
// entrypoint. Returns the build directory on success.
func (b *Builder) Build(ctx context.Context, recipe *manifest.Recipe, entrypoint []string, output string) (string, error) {
	stageImages := make(map[string]string)
	var currentCtr *compute.Container
	var finalSpec *aff.Spec

	for i := range recipe.Stages {
		if currentCtr != nil {
			currentCtr.Destroy(ctx)
		}
		ctr, spec, err := b.runStage(ctx, i+1, &recipe.Stages[i], stageImages)
		if err != nil {
			return "", err
		}
		currentCtr, finalSpec = ctr, spec
	}
	defer currentCtr.Destroy(ctx)

	if err := b.setEntrypoint(ctx, currentCtr, entrypoint); err != nil {
		return "", err
	}

	return b.exportImage(ctx, currentCtr, finalSpec, output)
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

// Exports the final image as an OCI tar archive and emits the affordance artifact.
//
// The image is exported to [files.ImageFile] within the output directory. The
// compiled affordance sections are written to [files.AffordanceFile] for the
// publish step to attach. When an image declares only OCI-level affordances
// the non-OCI sections are still emitted with the baseline. Returns the output
// directory path on success.
func (b *Builder) exportImage(ctx context.Context, ctr *compute.Container, spec *aff.Spec, output string) (string, error) {
	if err := os.MkdirAll(output, files.DefaultDirMode); err != nil {
		return "", crex.Wrap(oci.ErrFileSystemOperation, err)
	}

	image := filepath.Join(output, files.ImageFile)
	f, err := os.OpenFile(image, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, files.DefaultFileMode)
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

	if err := writeAffordance(spec, output); err != nil {
		return "", err
	}

	return output, nil
}

// Writes the compiled non-OCI affordance sections to [files.AffordanceFile].
//
// The OCI section is dropped because it ships to the runtime through the OCI
// config; the remaining sections are encoded. Always emitted so the runtime
// enforcement plugin receives the baseline for every service image.
func writeAffordance(spec *aff.Spec, output string) error {
	artifact := *spec
	artifact.OCI = nil
	payload, err := codec.Encode(&artifact, codec.JSON)
	if err != nil {
		return crex.Wrap(ErrBuild, err)
	}
	path := filepath.Join(output, files.AffordanceFile)
	if err := os.WriteFile(path, payload, files.DefaultFileMode); err != nil {
		return crex.Wrap(oci.ErrFileSystemOperation, err)
	}
	return nil
}
