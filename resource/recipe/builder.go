package recipe

import (
	"context"
	"io"

	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/spec/agent"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/crex"
)

// Author label recorded in the OCI history of the final exported image.
const commitAuthorExport = "export"

// Orchestrates a recipe build pipeline against a machine.
//
// Drives the stage pipeline: imports base images and executes each step in
// order. The final image is exported as an OCI tar archive.
type Builder struct {
	src     hub.Source  // Registry access for pulling base images.
	workdir string      // Manifest directory.
	agent   agent.Agent // Agent on the machine that runs the build.
}

// Returns a new Builder.
//
// src provides registry access for pulling base images. workdir is the
// directory containing the manifest and is the root for resolving copy step
// sources. a is the open agent on the machine that runs the build.
func NewBuilder(src hub.Source, workdir string, a agent.Agent) *Builder {
	return &Builder{
		src:     src,
		workdir: workdir,
		agent:   a,
	}
}

// Executes a recipe and writes the output image to w as an OCI tar archive.
//
// Each stage resolves its base image and executes its steps in order. The
// image produced by the last stage is the output. When set, entrypoint
// becomes the image entrypoint. The recipe must be valid, so it has at least
// one stage.
func (b *Builder) Build(ctx context.Context, recipe *manifest.Recipe, entrypoint []string, w io.Writer) error {
	if err := recipe.Validate(); err != nil {
		return err
	}

	stageImages := make(map[string]agent.ImageRef)
	var currentCtr agent.Container
	defer func() {
		if currentCtr != nil {
			currentCtr.Destroy(ctx)
		}
	}()

	for i := range recipe.Stages {
		if currentCtr != nil {
			currentCtr.Destroy(ctx)
			currentCtr = nil
		}
		ctr, err := b.runStage(ctx, i+1, &recipe.Stages[i], stageImages)
		if err != nil {
			return err
		}
		currentCtr = ctr
	}

	if err := b.setEntrypoint(ctx, currentCtr, entrypoint); err != nil {
		return err
	}

	return b.exportImage(ctx, currentCtr, w)
}

// Sets the image entrypoint on the container before it is committed.
//
// The entrypoint comes from the resource config rather than from a step, so
// it is applied as an image config update after all stages and steps have
// run.
func (b *Builder) setEntrypoint(ctx context.Context, ctr agent.Container, entrypoint []string) error {
	if len(entrypoint) == 0 {
		return nil
	}
	cfg, err := ctr.Inspect(ctx)
	if err != nil {
		return crex.Wrap(ErrBuild, err)
	}
	cfg.Entrypoint = entrypoint
	if err := ctr.Configure(ctx, cfg); err != nil {
		return crex.Wrap(ErrBuild, err)
	}
	return nil
}

// Commits the container and writes its image to w as an OCI tar archive.
func (b *Builder) exportImage(ctx context.Context, ctr agent.Container, w io.Writer) error {
	ref, err := ctr.Commit(ctx, commitAuthorExport)
	if err != nil {
		return crex.Wrap(ErrBuild, err)
	}
	if err := b.agent.Export(ctx, ref, w); err != nil {
		return crex.Wrap(ErrBuild, err)
	}
	return nil
}
