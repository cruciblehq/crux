package cli

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
)

// Represents the 'crux start' command.
type StartCmd struct{}

// Starts the resource.
func (c *StartCmd) Run(ctx context.Context) error {

	slog.Info("starting resource...")

	man, r, err := resource.Resolve(paths.Manifest(RootCmd.Context))
	if err != nil {
		return err
	}

	if err := r.Start(ctx, *man, paths.ImageTar(RootCmd.Context)); err != nil {
		if errors.Is(err, daemon.ErrConnectionRefused) {
			return crex.SystemError("daemon connection refused", err.Error()).
				Fallback("Wait a few seconds and try again. If the problem persists, try 'crux runtime restart' or 'crux runtime reset'.").
				Cause(err).
				Err()
		}
		return err
	}

	slog.Info("resource started")
	return nil
}
