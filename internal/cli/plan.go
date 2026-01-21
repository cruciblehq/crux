package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/pkg/plan"
)

// Represents the 'crux plan' command.
type PlanCmd struct {
	Blueprint string `arg:"" help:"Path to blueprint file"`
	State     string `optional:"" help:"Path to existing state file for incremental planning"`
}

// Executes the plan command.
func (c *PlanCmd) Run(ctx context.Context) error {
	opts := plan.Options{
		Blueprint: c.Blueprint,
		State:     c.State,
	}

	slog.Info("generating deployment plan...", "blueprint", c.Blueprint, "state", c.State)

	result, err := plan.Plan(ctx, opts)
	if err != nil {
		return err
	}

	slog.Info("deployment plan generated successfully", "output", result.Output)

	return nil
}
