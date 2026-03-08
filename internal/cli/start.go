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
	client, err := b.Client(ctx, internal.InstanceName)
	if err != nil {
		return err
	}

	if err := containerStart(ctx, client, *man, paths.ImageTar(RootCmd.Context)); err != nil {
		if errors.Is(err, compute.ErrConnectionRefused) {
			return crex.SystemError("daemon connection refused", err.Error()).
				Fallback("Wait a few seconds and try again. If the problem persists, try 'crux runtime restart' or 'crux runtime reset'.").
				Cause(err).
				Err()
		}
		return err
	}

	slog.Info("resource started")
	return nil
}

// Ensures a container is running for the given resource.
//
// The operation is idempotent: if the container is already running it is
// left untouched; if it is stopped a new process is started on the
// existing snapshot; if it does not exist the image is imported and a
// fresh container is created.
func containerStart(ctx context.Context, client compute.Client, m manifest.Manifest, path string) error {
	result, err := client.ContainerStatus(ctx, &protocol.ContainerStatusRequest{ID: m.Resource.Name})
	if err != nil {
		return err
	}

	switch result.Status {
	case protocol.ContainerRunning:
		return nil

	case protocol.ContainerStopped:
		return client.ImageStart(ctx, &protocol.ImageStartRequest{
			Ref:     m.Resource.Name,
			Version: m.Resource.Version,
		})

	default:
		if err := client.ImageImport(ctx, &protocol.ImageImportRequest{
			Ref:     m.Resource.Name,
			Version: m.Resource.Version,
			Path:    path,
		}); err != nil {
			return err
		}

		return client.ImageStart(ctx, &protocol.ImageStartRequest{
			Ref:     m.Resource.Name,
			Version: m.Resource.Version,
		})
	}
}
