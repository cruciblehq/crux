package cli

import (
	"context"

	"github.com/cruciblehq/crux/pkg/plan"
)

// Generates a deployment plan from a blueprint.
type PlanCmd struct {
	Blueprint string `arg:"" help:"Path to blueprint file"`
	State     string `optional:"" help:"Path to existing state file for incremental planning"`
	Output    string `optional:"" help:"Output path for plan file (default: dist/plans/plan-<timestamp>.json)"`
}

// Executes the plan command.
func (c *PlanCmd) Run(ctx context.Context) error {
	opts := plan.PlanOptions{
		BlueprintPath: c.Blueprint,
		StatePath:     c.State,
		OutputPath:    c.Output,
	}

	_, _, err := plan.Generate(ctx, opts)
	if err != nil {
		return err
	}

	return nil
}
