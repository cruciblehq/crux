package service

import (
	"context"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/crux/resource/recipe"
	"github.com/cruciblehq/spec/manifest"
)

// Builds a Crucible service resource from its configuration.
//
// Drives the shared recipe pipeline against the local machine, wiring the
// configured entrypoint into the produced image.
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

// Builds the service described by cfg into the output directory.
//
// Connects to the agent on the local machine and runs the recipe with the
// service entrypoint. Returns the build directory on success.
func (b *Builder) Build(ctx context.Context, cfg *manifest.Service, output string) (string, error) {
	provider, err := compute.ProviderFor(compute.Local)
	if err != nil {
		return "", err
	}
	client, err := provider.Connect(ctx, compute.DefaultLocalMachineID())
	if err != nil {
		return "", err
	}
	defer client.Close()

	return recipe.NewBuilder(b.src, b.workdir, client).Build(ctx, &cfg.Recipe, cfg.Entrypoint, output)
}
