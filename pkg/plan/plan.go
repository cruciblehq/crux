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
	"github.com/cruciblehq/protocol/pkg/plan"
	"github.com/cruciblehq/protocol/pkg/state"
)

// Options for generating a deployment plan.
type Options struct {
	Blueprint string // Path to blueprint file.
	State     string // Path to existing state for incremental planning (optional).
	Output    string // Output path for plan file (optional).
}

// Result of generating a deployment plan.
type Result struct {
	Plan   *plan.Plan // The generated plan.
	Output string     // Path where the plan was written.
}

// Generates a deployment plan from a blueprint.
func Plan(ctx context.Context, opts Options) (*Result, error) {

	// Read blueprint
	bp, err := blueprint.Read(opts.Blueprint)
	if err != nil {
		return nil, crex.UserError("invalid blueprint", err.Error()).
			Fallback("Ensure the blueprint file exists and is valid.").
			Err()
	}

	// Read existing state if provided
	var st *state.State
	if opts.State != "" {
		st, err = state.Read(opts.State)
		if err != nil {
			return nil, crex.UserError("invalid state file", err.Error()).
				Fallback("Ensure the state file exists and is valid.").
				Err()
		}
	}

	// Generate plan
	p, err := build(ctx, bp, opts.Blueprint, st)
	if err != nil {
		return nil, err
	}

	// Determine output path
	output, err := determineOutputPath(opts.Output, opts.Blueprint)
	if err != nil {
		return nil, err
	}

	// Write plan
	if err := p.Write(output); err != nil {
		return nil, err
	}

	return &Result{
		Plan:   p,
		Output: output,
	}, nil
}

// Determines the output path for the plan file.
// If outputPath is provided, it is used. Otherwise, a timestamped path is generated.
func determineOutputPath(outputPath, blueprintPath string) (string, error) {
	if outputPath != "" {
		return outputPath, nil
	}

	timestamp := time.Now().Format("20060102-150405")
	dir := filepath.Dir(blueprintPath)
	plansDir := filepath.Join(dir, "dist", "plans")
	if err := os.MkdirAll(plansDir, paths.DefaultDirMode); err != nil {
		return "", err
	}
	return filepath.Join(plansDir, fmt.Sprintf("plan-%s.json", timestamp)), nil
}

// Generates a plan from a blueprint.
// If state is provided, generates an incremental plan based on current deployment state.
func build(ctx context.Context, bp *blueprint.Blueprint, blueprintPath string, st *state.State) (*plan.Plan, error) {
	p := &plan.Plan{
		Version:  1,
		Services: make([]plan.Service, 0, len(bp.Services)),
		Gateway: plan.Gateway{
			Routes: make([]plan.Route, 0, len(bp.Services)),
		},
		Infrastructure: plan.Infrastructure{
			Provider: "local", // TODO: Determine provider from config
		},
	}

	// For now, just create a simple plan for local Docker deployment
	// TODO: Actually resolve references, fetch manifests, validate, etc.

	for _, svc := range bp.Services {
		// TODO: Parse reference string into reference.Reference
		// TODO: Resolve version constraint to exact version + digest
		// For now, create a placeholder service with empty Reference
		service := plan.Service{
			ID: svc.ID, // Use ID from blueprint
			// Reference: will be populated when resolution is implemented
		}
		p.Services = append(p.Services, service)

		// Create gateway route
		route := plan.Route{
			Pattern:   svc.Prefix + "/*", // Add wildcard for all sub-paths
			ServiceID: svc.ID,
		}
		p.Gateway.Routes = append(p.Gateway.Routes, route)
	}

	return p, nil
}
