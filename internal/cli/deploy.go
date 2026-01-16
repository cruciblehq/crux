package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/pkg/deploy"
)

// Executes a deployment plan.
type DeployCmd struct {
	Plan     string `arg:"" help:"Path to plan file"`
	State    string `optional:"" help:"Path to existing state file for incremental deployment"`
	Output   string `optional:"" help:"Output path for state file (default: dist/states/state-<timestamp>.json)"`
	Provider string `optional:"" help:"Provider configuration name (uses default if not specified)"`
}

// Executes the deploy command.
func (c *DeployCmd) Run(ctx context.Context) error {
	opts := deploy.DeployOptions{
		PlanPath:     c.Plan,
		StatePath:    c.State,
		OutputPath:   c.Output,
		ProviderName: c.Provider,
	}

	slog.Info("starting deployment")

	result, err := deploy.Execute(ctx, opts)
	if err != nil {
		return err
	}

	slog.Info("deployment completed successfully", "state", result.OutputPath)

	return nil
}
