package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/crux/resource/blueprint"
	"github.com/cruciblehq/spec/registry"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Represents the 'crux local deploy' command.
type LocalDeployCmd struct {
	Registry    string `help:"Hub registry URL for fetching runtimes (default: http://hub.cruciblehq.xyz:8080)."`
	Environment string `short:"e" help:"Blueprint environment to deploy." default:""`
}

// Builds the local blueprint into a deployment plan and writes it to disk.
//
// Writes the resulting plan to the local build directory. The plan records
// what the environment must look like; actual container lifecycle management is
// a separate concern.
func (c *LocalDeployCmd) Run(ctx context.Context) error {
	slog.Info("building local deployment plan...")

	bp, err := localBlueprint()
	if err != nil {
		return err
	}

	registryURL := c.Registry
	if registryURL == "" {
		registryURL = internal.DefaultRegistryURL
	}
	src, err := hub.NewSource(registryURL, internal.DefaultNamespace)
	if err != nil {
		return err
	}

	output := file.BuildDir(localStateDir())
	if err := os.MkdirAll(output, file.DefaultDirMode); err != nil {
		return crex.SystemError("cannot prepare build output", "failed to create the build output directory").
			Recoveryf("Make sure you have write access to %s, then try again.", output).
			Cause(crex.Wrap(registry.ErrFileSystemOperation, err)).
			Err()
	}

	if err := blueprint.NewBuilder(src, c.Environment).Build(ctx, bp, output); err != nil {
		return err
	}

	slog.Info("local plan written", "path", file.Plan(output))
	return nil
}
