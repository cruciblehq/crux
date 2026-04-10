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

// Represents the 'crux reset' command.
type ResetCmd struct{}

// Destroys the resource and recreates it.
func (c *ResetCmd) Run(ctx context.Context) error {

	slog.Info("resetting resource...")

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

	tag := runtime.ImageTag(man.Resource.Name, man.Resource.Version)
	rt.DestroyImage(ctx, tag)

	if err := containerStart(ctx, rt, *man, paths.ImageTar(RootCmd.Context)); err != nil {
		return err
	}

	slog.Info("resource reset complete")
	return nil
}
