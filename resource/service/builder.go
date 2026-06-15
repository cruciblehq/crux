package service

import (
	"context"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/resource/recipe"
	"github.com/cruciblehq/crux/source"
)

// Builds a Crucible service resource from its configuration.
//
// Drives the shared recipe pipeline against the local compute backend, wiring
// the configured entrypoint into the produced image.
type Builder struct {
	src     source.Source // Registry access for pulling base images and resolving affordances.
	workdir string        // Manifest directory, used as the root for resolving copy step sources.
}

// Returns a new Builder.
//
// source provides registry access for pulling base images and resolving
// affordance references. workdir is the directory containing the manifest and
// is the root for resolving copy step sources.
func NewBuilder(src source.Source, workdir string) *Builder {
	return &Builder{src: src, workdir: workdir}
}

// Builds the service described by cfg into the output directory.
//
// Connects to the local compute backend and runs the recipe with the service
// entrypoint. Returns the build directory on success.
func (b *Builder) Build(ctx context.Context, cfg *manifest.Service, output string) (string, error) {
	backend, err := compute.BackendFor(compute.Local)
	if err != nil {
		return "", err
	}
	client, err := backend.Connect(ctx, compute.LocalInstance)
	if err != nil {
		return "", err
	}
	defer client.Close()

	return recipe.NewBuilder(b.src, b.workdir, client).Build(ctx, &cfg.Recipe, cfg.Entrypoint, output)
}
