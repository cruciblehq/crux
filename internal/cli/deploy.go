package cli

import (
	"fmt"

	"github.com/cruciblehq/crux/pkg/plan"
	"github.com/cruciblehq/protocol/pkg/blueprint"
)

// Deploys a blueprint using a plan.
type DeployCmd struct {
	Blueprint string `arg:"" help:"Path to blueprint file"`
	Plan      string `optional:"" help:"Path to plan file (if omitted, will auto-generate plan)"`
	Target    string `optional:"" help:"Target environment (required if --plan not provided)"`
}

// Executes the deploy command.
func (c *DeployCmd) Run() error {
	var p *plan.Plan
	var err error

	// Load or generate plan
	if c.Plan != "" {
		// Use existing plan
		p, err = plan.Read(c.Plan)
		if err != nil {
			return fmt.Errorf("failed to read plan: %w", err)
		}
		fmt.Printf("Using plan: %s\n", c.Plan)
	} else {
		// Auto-generate plan
		if c.Target == "" {
			return fmt.Errorf("--target is required when --plan is not provided")
		}

		bp, err := blueprint.Read(c.Blueprint)
		if err != nil {
			return fmt.Errorf("failed to read blueprint: %w", err)
		}

		p, err = plan.Build(bp, c.Target, c.Blueprint)
		if err != nil {
			return fmt.Errorf("failed to build plan: %w", err)
		}

		fmt.Printf("Auto-generated plan for target: %s\n", c.Target)
	}

	// Execute deployment
	fmt.Printf("\nDeploying...\n")
	fmt.Printf("Target: %s\n", p.Target)
	fmt.Printf("Services to deploy: %d\n", len(p.Services))

	for _, svc := range p.Services {
		fmt.Printf("  - %s at %s\n", svc.Reference, svc.Prefix)
	}

	// TODO: Actually deploy (load Docker images, start containers, configure gateway)
	fmt.Println("\n⚠️  Deployment not yet implemented - plan validated successfully")

	return nil
}
