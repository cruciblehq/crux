package cli

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
	"github.com/cruciblehq/crux/internal/manifest"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
	"github.com/cruciblehq/crux/internal/runtime"
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
	rt, err := b.Runtime(ctx, internal.DefaultInstanceName)
	if err != nil {
		return err
	}
	defer rt.Close()

	name := runtime.ContainerID(man.Resource.Name)
	state, err := rt.Container(name).Status(ctx)
	if err != nil {
		return err
	}

	fmt.Println(state)
	return nil
}
