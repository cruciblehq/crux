package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/pkg/pack"
)

// Represents the 'crux pack' command
type PackCmd struct{}

// Executes the pack command
func (c *PackCmd) Run(ctx context.Context) error {

	// Package the built resource
	if err := pack.Pack(ctx); err != nil {
		return err
	}

	slog.Info("package created successfully", "path", pack.PackageOutput)

	return nil
}
