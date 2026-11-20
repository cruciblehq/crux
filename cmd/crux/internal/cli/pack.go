package cli

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/utils-go/crex"
)

// Represents the 'crux pack' command.
type PackCmd struct{}

// Executes the pack command.
//
// Packages the built resource into an archive at the package path. The build
// step must have completed first, or an error is returned. On success, logs
// the path of the created package.
func (c *PackCmd) Run(ctx context.Context) error {

	slog.Info("packaging resource...", "output", resource.PackagePath(RootCmd.Context))

	result, err := resource.Pack(ctx, resource.BuildDir(RootCmd.Context), resource.PackagePath(RootCmd.Context))
	if err != nil {
		if errors.Is(err, resource.ErrFileSystemOperation) {
			return crex.SystemError("cannot package resource", "the package output directory could not be created").
				Recoveryf("Make sure you have write access to %s, then try again.", resource.PackagePath(RootCmd.Context)).
				Reclassify(err)
		}
		return err
	}

	slog.Info("package created successfully", "path", result.Output)

	return nil
}
