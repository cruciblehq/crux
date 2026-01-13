package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cruciblehq/crux/pkg/plan"
	"github.com/cruciblehq/protocol/pkg/blueprint"
)

// Generates a deployment plan from a blueprint.
type PlanCmd struct {
	Blueprint string `arg:"" help:"Path to blueprint file"`
	Target    string `required:"" help:"Target environment name (e.g., local, aws-production)"`
	Output    string `optional:"" help:"Output path for plan file (default: dist/plans/plan-<target>-<timestamp>.json)"`
}

// Executes the plan command.
func (c *PlanCmd) Run() error {
	// Read blueprint
	bp, err := blueprint.Read(c.Blueprint)
	if err != nil {
		return fmt.Errorf("failed to read blueprint: %w", err)
	}

	// Generate plan
	p, err := plan.Build(bp, c.Target, c.Blueprint)
	if err != nil {
		return fmt.Errorf("failed to build plan: %w", err)
	}

	// Determine output path
	outputPath := c.Output
	if outputPath == "" {
		timestamp := time.Now().Format("20060102-150405")
		dir := filepath.Dir(c.Blueprint)
		plansDir := filepath.Join(dir, "dist", "plans")
		if err := os.MkdirAll(plansDir, 0755); err != nil {
			return fmt.Errorf("failed to create plans directory: %w", err)
		}
		outputPath = filepath.Join(plansDir, fmt.Sprintf("plan-%s-%s.json", c.Target, timestamp))
	}

	// Write plan
	if err := p.Write(outputPath); err != nil {
		return fmt.Errorf("failed to write plan: %w", err)
	}

	fmt.Printf("Plan created: %s\n", outputPath)
	fmt.Printf("Target: %s\n", p.Target)
	fmt.Printf("Services: %d\n", len(p.Services))

	return nil
}
