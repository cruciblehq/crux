package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/protocol"
)

// Represents the 'crux status' command.
type StatusCmd struct{}

// Shows the current state of the resource.
func (c *StatusCmd) Run(ctx context.Context) error {
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

	result, err := client.ContainerStatus(ctx, &protocol.ContainerStatusRequest{
		ID: man.Resource.Name,
	})
	if err != nil {
		if errors.Is(err, compute.ErrConnectionRefused) {
			return crex.SystemError("daemon connection refused", err.Error()).
				Fallback("Wait a few seconds and try again. If the problem persists, try 'crux runtime restart' or 'crux runtime reset'.").
				Cause(err).
				Err()
		}
		return err
	}

	fmt.Println(result.Status)
	return nil
}
