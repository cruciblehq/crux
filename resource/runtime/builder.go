package runtime

import (
	"context"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/crux/resource/recipe"
	"github.com/cruciblehq/spec/manifest"
)

// Builds a Crucible runtime resource from its configuration.
//
// Drives the shared recipe pipeline against the local compute backend. A
// runtime has no entrypoint, so the produced image keeps the entrypoint of its
// final stage.
type Builder struct {
	src     hub.Source // Registry access for pulling base images and resolving affordances.
	workdir string     // Manifest directory, used as the root for resolving copy step sources.
}

// Returns a new Builder.
//
// source provides registry access for pulling base images and resolving
// affordance references. workdir is the directory containing the manifest and
// is the root for resolving copy step sources.
func NewBuilder(src hub.Source, workdir string) *Builder {
	return &Builder{src: src, workdir: workdir}
}

// Builds the runtime described by cfg into the output directory.
//
// Connects to the local compute backend and runs the recipe. Returns the build
// directory on success.
func (b *Builder) Build(ctx context.Context, cfg *manifest.Runtime, output string) (string, error) {
	backend, err := compute.BackendFor(compute.Local)
	if err != nil {
		return "", err
	}
	client, err := backend.Connect(ctx, compute.LocalInstance)
	if err != nil {
		return "", err
	}
	defer client.Close()

	return recipe.NewBuilder(b.src, b.workdir, client).Build(ctx, &cfg.Recipe, nil, output)
}
