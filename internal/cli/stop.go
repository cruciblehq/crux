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

// Represents the 'crux stop' command.
type StopCmd struct{}

// Stops the resource.
func (c *StopCmd) Run(ctx context.Context) error {

	slog.Info("stopping resource...")

	man, r, err := resource.Resolve(paths.Manifest(RootCmd.Context))
	if err != nil {
		return err
	}

	if err := r.Stop(ctx, *man); err != nil {
		if errors.Is(err, daemon.ErrConnectionRefused) {
			return crex.SystemError("daemon connection refused", err.Error()).
				Fallback("Wait a few seconds and try again. If the problem persists, try 'crux runtime restart' or 'crux runtime reset'.").
				Cause(err).
				Err()
		}
		return err
	}

	slog.Info("resource stopped")
	return nil
}
