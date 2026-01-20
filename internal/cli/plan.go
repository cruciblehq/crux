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
	opts := plan.PlanOptions{
		BlueprintPath: c.Blueprint,
		StatePath:     c.State,
	}

	slog.Info("generating deployment plan...", "blueprint", c.Blueprint, "state", c.State)

	_, _, err := plan.Generate(ctx, opts)
	if err != nil {
		return err
	}

	slog.Info("deployment plan generated successfully")

	return nil
}
