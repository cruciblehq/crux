package cli

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/spec/registry"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Represents the 'crux pack' command.
type PackCmd struct{}

// Executes the pack command.
//
// Packages the built resources into a deployable artifact. The command assumes
// that the build step has already been completed successfully. If not, it will
// return an error. Upon successful packaging, it logs the output path of the
// created package.
func (c *PackCmd) Run(ctx context.Context) error {

	slog.Info("packaging resource...", "output", file.Package(RootCmd.Context))

	result, err := resource.Pack(ctx, file.BuildDir(RootCmd.Context), file.Package(RootCmd.Context))
	if err != nil {
		if errors.Is(err, registry.ErrFileSystemOperation) {
			return crex.SystemError("cannot package resource", "the package output directory could not be created").
				Recoveryf("Make sure you have write access to %s, then try again.", file.Package(RootCmd.Context)).
				Reclassify(err)
		}
		return err
	}

	slog.Info("package created successfully", "path", result.Output)

	return nil
}
