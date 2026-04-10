package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
	"github.com/cruciblehq/crux/internal/manifest"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
	"github.com/cruciblehq/crux/internal/runtime"
)

// Represents the 'crux start' command.
type StartCmd struct{}

// Starts the resource.
func (c *StartCmd) Run(ctx context.Context) error {

	slog.Info("starting resource...")

	man, err := resource.ReadManifest(paths.Manifest(RootCmd.Context))
	if err != nil {
		return err
	}

	if man.Resource.Type != manifest.TypeService {
		return resource.ErrUnsupported
	}

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	rt, err := b.Runtime(ctx, internal.DefaultInstanceName)
	if err != nil {
		return err
	}
	defer rt.Close()

	if err := containerStart(ctx, rt, *man, paths.ImageTar(RootCmd.Context)); err != nil {
		return err
	}

	slog.Info("resource started")
	return nil
}

// Ensures a container is running for the given resource.
//
// The image is imported if not already present, then [Runtime.StartFromTag]
// handles the container lifecycle idempotently: if the container is already
// running it is left untouched; if it is stopped a new task is started on
// the existing snapshot; if it does not exist a fresh container is created.
func containerStart(ctx context.Context, rt *runtime.Runtime, m manifest.Manifest, path string) error {
	cfg := m.Config.(*manifest.Service)

	tag := runtime.ImageTag(m.Resource.Name, m.Resource.Version)
	name := runtime.ContainerID(m.Resource.Name)

	if err := rt.ImportImage(ctx, path, tag); err != nil {
		return err
	}

	outputStage := cfg.Stages[len(cfg.Stages)-1]

	_, err := rt.StartFromTag(ctx, tag, name, outputStage.Affordances)
	return err
}
