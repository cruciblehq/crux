package runtime

import (
	"context"
	"io"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/crux/resource/recipe"
	"github.com/cruciblehq/spec/manifest"
)

// Builds a Crucible runtime resource from its configuration.
//
// Drives the shared recipe pipeline against the local machine. A runtime has
// no entrypoint, so the produced image keeps the entrypoint of its final
// stage.
type Builder struct {
	src     hub.Source // Registry access for pulling base images.
	workdir string     // Manifest directory, used as the root for resolving copy step sources.
}

// Returns a new Builder.
//
// src provides registry access for pulling base images. workdir is the
// directory containing the manifest and is the root for resolving copy step
// sources.
func NewBuilder(src hub.Source, workdir string) *Builder {
	return &Builder{src: src, workdir: workdir}
}

// Builds the runtime described by cfg and writes its image to w.
//
// Connects to the agent on the local machine and runs the recipe. The image
// is an OCI tar archive.
func (b *Builder) Build(ctx context.Context, cfg *manifest.Runtime, w io.Writer) error {
	provider, err := compute.ProviderFor(compute.Local)
	if err != nil {
		return err
	}
	client, err := provider.Connect(ctx, compute.DefaultLocalMachineID())
	if err != nil {
		return err
	}
	defer client.Close()

	return recipe.NewBuilder(b.src, b.workdir, client).Build(ctx, &cfg.Recipe, nil, w)
}
