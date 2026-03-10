package cli

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/protocol"
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
	client, err := b.Client(ctx, internal.DefaultInstanceName)
	if err != nil {
		return err
	}

	if err := client.ContainerDestroy(ctx, &protocol.ContainerDestroyRequest{
		ID: man.Resource.Name,
	}); err != nil {
		if errors.Is(err, compute.ErrConnectionRefused) {
			return crex.SystemError("daemon connection refused", err.Error()).
				Fallback("Wait a few seconds and try again. If the problem persists, try 'crux runtime restart' or 'crux runtime reset'.").
				Cause(err).
				Err()
		}
		return err
	}

	slog.Info("resource destroyed")
	return nil
}
