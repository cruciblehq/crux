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

// Represents the 'crux destroy' command.
type DestroyCmd struct{}

// Destroys the resource.
func (c *DestroyCmd) Run(ctx context.Context) error {

	slog.Info("destroying resource...")

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
	if err := rt.DestroyImage(ctx, tag); err != nil {
		return err
	}

	slog.Info("resource destroyed")
	return nil
}
