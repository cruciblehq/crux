package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/files"
	"github.com/cruciblehq/crux/resource/blueprint"
	"github.com/cruciblehq/crux/source"
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

	registry := c.Registry
	if registry == "" {
		registry = internal.DefaultRegistryURL
	}
	src, err := source.NewSource(registry, internal.DefaultNamespace)
	if err != nil {
		return err
	}

	output := files.BuildDir(files.LocalDir())
	if err := os.MkdirAll(output, files.DefaultDirMode); err != nil {
		return err
	}

	if err := blueprint.NewBuilder(src, c.Environment).Build(ctx, bp, output); err != nil {
		return err
	}

	slog.Info("local plan written", "path", files.Plan(output))
	return nil
}
