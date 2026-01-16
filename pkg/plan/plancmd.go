package plan

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/paths"
	"github.com/cruciblehq/protocol/pkg/blueprint"
)

// Options for generating a deployment plan.
type PlanOptions struct {
	BlueprintPath string // Path to blueprint file
	StatePath     string // Optional path to existing state for incremental planning
	OutputPath    string // Optional output path for plan file
}

// Generate generates a deployment plan from a blueprint.
//
// Returns the created plan and the output path where it was written.
func Generate(ctx context.Context, opts PlanOptions) (*Plan, string, error) {
	// Read blueprint
	bp, err := blueprint.Read(opts.BlueprintPath)
	if err != nil {
		return nil, "", crex.UserError("invalid blueprint", err.Error()).
			Fallback("Ensure the blueprint file exists and is valid YAML.").
			Err()
	}

	// Read existing state if provided
	var state *State
	if opts.StatePath != "" {
		state, err = ReadState(opts.StatePath)
		if err != nil {
			return nil, "", crex.UserError("invalid state file", err.Error()).
				Fallback("Ensure the state file exists and is valid JSON.").
				Err()
		}
	}

	// Generate plan
	p, err := Build(ctx, bp, opts.BlueprintPath, state)
	if err != nil {
		return nil, "", err
	}

	// Determine output path
	outputPath := opts.OutputPath
	if outputPath == "" {
		timestamp := time.Now().Format("20060102-150405")
		dir := filepath.Dir(opts.BlueprintPath)
		plansDir := filepath.Join(dir, "dist", "plans")
		if err := os.MkdirAll(plansDir, paths.DefaultDirMode); err != nil {
			return nil, "", err
		}
		outputPath = filepath.Join(plansDir, fmt.Sprintf("plan-%s.json", timestamp))
	}

	// Write plan
	if err := p.Write(outputPath); err != nil {
		return nil, "", err
	}

	return p, outputPath, nil
}
