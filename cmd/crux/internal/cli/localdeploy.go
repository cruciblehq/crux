package cli

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/dirs"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Represents the 'crux local deploy' command.
type LocalDeployCmd struct {
	Registry    string `help:"Registry to fetch resources from (default: the one set with 'crux registry use')."`
	Environment string `short:"e" help:"Blueprint environment to deploy." default:""`
}

// Compiles the local blueprint into a deployment plan and writes it to disk.
//
// Writes the plan to the local build directory. The plan records what the
// environment must look like. Applying it to the machine is a separate step.
func (c *LocalDeployCmd) Run(ctx context.Context) error {
	slog.Info("building local deployment plan...")

	m, _, err := openLocalBlueprint()
	if err != nil {
		return err
	}

	src, err := openRegistrySource(c.Registry)
	if err != nil {
		return err
	}

	output := resource.BuildDir(dirs.Local())
	if err := os.MkdirAll(output, file.DefaultDirMode); err != nil {
		return crex.SystemError("cannot prepare build output", "failed to create the build output directory").
			Recoveryf("Make sure you have write access to %s, then try again.", output).
			Cause(crex.Wrap(resource.ErrFileSystemOperation, err)).
			Err()
	}

	if _, err := resource.Build(ctx, *m, src, RootCmd.Context, c.Environment, output); err != nil {
		return err
	}

	slog.Info("local plan written", "path", filepath.Join(output, manifest.PlanFile))
	return nil
}
