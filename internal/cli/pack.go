package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal/pack"
	"github.com/cruciblehq/crux/internal/paths"
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

	slog.Info("packaging resource...", "output", paths.Package(RootCmd.Context))

	// Package the built resource
	result, err := pack.Pack(ctx, pack.Options{
		Manifest: paths.Manifest(RootCmd.Context),
		Dist:     paths.BuildDir(RootCmd.Context),
		Output:   paths.Package(RootCmd.Context),
	})
	if err != nil {
		return err
	}

	slog.Info("package created successfully", "path", result.Output)

	return nil
}
