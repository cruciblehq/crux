package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/pkg/deploy"
)

// Represents the 'crux deploy' command.
type DeployCmd struct {
	Plan     string `arg:"" help:"Path to plan file"`
	State    string `optional:"" help:"Path to existing state file for incremental deployment"`
	Provider string `optional:"" help:"Provider configuration name (uses default if not specified)"`
}

// Executes the deploy command.
func (c *DeployCmd) Run(ctx context.Context) error {
	opts := deploy.DeployOptions{
		PlanPath:     c.Plan,
		StatePath:    c.State,
		ProviderName: c.Provider,
	}

	slog.Info("deploying plan...", "plan", c.Plan, "state", c.State, "provider", c.Provider)

	result, err := deploy.Execute(ctx, opts)
	if err != nil {
		return err
	}

	slog.Info("deployment completed successfully", "state", result.OutputPath)

	return nil
}
