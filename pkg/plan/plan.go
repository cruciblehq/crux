package plan

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cruciblehq/crux/pkg/config"
	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/paths"
	"github.com/cruciblehq/protocol/pkg/blueprint"
	"github.com/cruciblehq/protocol/pkg/plan"
	"github.com/cruciblehq/protocol/pkg/registry"
	"github.com/cruciblehq/protocol/pkg/state"
)

const (

	// Directory for plan outputs relative to blueprint location.
	planOutputDir = "dist/plans"

	// Timestamp format for plan filenames.
	timestampFormat = "20060102-150405"

	// Current supported plan version.
	Version = 1

	// Default compute instance type for all deployments.
	DefaultInstanceType = "t3.micro"
)

// Options for generating a deployment plan.
type Options struct {
	Blueprint string // Path to blueprint file.
	State     string // Path to existing state for incremental planning (optional).
	Output    string // Output path for plan file (optional).
	Registry  string // Registry URL for resolving references.
	Provider  string // Provider profile name (empty = default).
}

// Result of generating a deployment plan.
type Result struct {
	Plan   *plan.Plan // The generated plan.
	Output string     // Path where the plan was written.
}

// Generates a deployment plan from a blueprint.
//
// If state is provided, generates an incremental plan based on current
// deployment state. The generated plan is written to the specified output path
// or to a default location if no output path is provided.
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

	// Load provider configuration
	provider, err := config.GetOrDefaultProvider(opts.Provider)
	if err != nil {
		return nil, crex.UserError("provider not found", err.Error()).
			Fallback("Run 'crux provider list' to see configured providers or 'crux provider add' to add one.").
			Err()
	}

	// Generate plan
	p, err := build(ctx, bp, st, opts.Registry, provider.Type)
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
//
// If outputPath is provided, it is used. Otherwise, a path based on the current
// timestamp and blueprint location is generated.
func determineOutputPath(outputPath, blueprintPath string) (string, error) {
	if outputPath != "" {
		return outputPath, nil
	}

	timestamp := time.Now().Format(timestampFormat)
	dir := filepath.Dir(blueprintPath)
	plansDir := filepath.Join(dir, planOutputDir)
	if err := os.MkdirAll(plansDir, paths.DefaultDirMode); err != nil {
		return "", err
	}
	return filepath.Join(plansDir, fmt.Sprintf("plan-%s%s", timestamp, ".json")), nil
}

// Generates a plan from a blueprint.
//
// Transforms a blueprint containing symbolic service references into a concrete
// deployment plan with frozen references. Groups services by namespace and name,
// intersects version constraints when multiple instances reference the same service,
// and resolves each reference to a specific version with digest by querying the
// registry. When state is provided, the function compares against the current
// deployment to enable incremental planning.
func build(ctx context.Context, bp *blueprint.Blueprint, st *state.State, registryURL string, providerType config.ProviderType) (*plan.Plan, error) {
	p := &plan.Plan{
		Version:      Version,
		Services:     make([]plan.Service, 0, len(bp.Services)),
		Compute:      make([]plan.Compute, 0, 1),
		Environments: make([]plan.Environment, 0),
		Bindings:     make([]plan.Binding, 0, len(bp.Services)),
		Gateway: plan.Gateway{
			Routes: make([]plan.Route, 0, len(bp.Services)),
		},
	}

	registryClient := registry.NewClient(registryURL, nil)

	// Resolve all service references
	if err := resolveServiceReferences(ctx, bp, st, registryClient, p); err != nil {
		return nil, err
	}

	// Allocate compute resources
	allocateCompute(p, string(providerType))

	// Create bindings between services and compute
	bind(p)

	return p, nil
}
