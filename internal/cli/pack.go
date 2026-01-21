package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/pkg/pack"
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

	slog.Info("packaging resource...", "output", Package)

	// Package the built resource
	result, err := pack.Pack(ctx, pack.Options{
		Manifest: Manifestfile,
		Dist:     Dist,
		Output:   Package,
	})
	if err != nil {
		return err
	}

	slog.Info("package created successfully", "path", result.Output)

	return nil
}
