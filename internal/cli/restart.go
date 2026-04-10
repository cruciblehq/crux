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

// Represents the 'crux restart' command.
type RestartCmd struct{}

// Restarts the resource, preserving its filesystem state.
func (c *RestartCmd) Run(ctx context.Context) error {

	slog.Info("restarting resource...")

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

	name := runtime.ContainerID(man.Resource.Name)
	rt.Container(name).Stop(ctx)

	if err := containerStart(ctx, rt, *man, paths.ImageTar(RootCmd.Context)); err != nil {
		return err
	}

	slog.Info("resource restarted")
	return nil
}
