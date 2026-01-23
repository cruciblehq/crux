package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/pkg/plan"
)

// Represents the 'crux plan' command.
type PlanCmd struct {
	Blueprint string `arg:"" help:"Path to blueprint file"`
	State     string `optional:"" help:"Path to existing state file for incremental planning"`
	Registry  string `help:"Registry URL for resolving references (default: http://hub.cruciblehq.xyz:8080)."`
	Provider  string `help:"Provider profile name (empty = default)"`
}

// Executes the plan command.
func (c *PlanCmd) Run(ctx context.Context) error {
	registry := c.Registry
	if registry == "" {
		registry = internal.DefaultRegistryURL
	}

	opts := plan.Options{
		Blueprint: c.Blueprint,
		State:     c.State,
		Registry:  registry,
		Provider:  c.Provider,
	}

	slog.Info("generating deployment plan...", "blueprint", c.Blueprint, "state", c.State)

	result, err := plan.Plan(ctx, opts)
	if err != nil {
		return err
	}

	slog.Info("deployment plan generated successfully", "output", result.Output)

	return nil
}
