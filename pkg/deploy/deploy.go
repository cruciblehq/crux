package deploy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cruciblehq/crux/pkg/config"
	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/paths"
	"github.com/cruciblehq/protocol/pkg/plan"
	"github.com/cruciblehq/protocol/pkg/state"
)

// Options for executing a deployment.
type Options struct {
	Plan     string // Path to plan file.
	State    string // Optional path to existing state for incremental deployment.
	Output   string // Optional output path for state file.
	Provider string // Optional provider configuration name (uses default if empty).
}

// Result of a deployment execution.
type Result struct {
	State  *state.State // New state after deployment.
	Output string       // Path where state was saved.
}

// Deploy executes a deployment plan.
func Deploy(ctx context.Context, opts Options) (*Result, error) {
	// Load provider configuration
	providerConfig, err := loadProviderConfig(opts.Provider)
	if err != nil {
		return nil, err
	}

	// Read plan
	p, err := plan.Read(opts.Plan)
	if err != nil {
		return nil, crex.Wrap(ErrInvalidPlan, err)
	}

	// Read existing state if provided
	var currentState *state.State
	if opts.State != "" {
		currentState, err = state.Read(opts.State)
		if err != nil {
			return nil, crex.Wrap(ErrInvalidState, err)
		}
	}

	// Create deployer
	deployer, err := createDeployer(providerConfig)
	if err != nil {
		return nil, err
	}

	// Execute deployment
	newState, err := deployer.Deploy(ctx, p, currentState)
	if err != nil {
		return nil, err
	}

	// Determine output path
	output := opts.Output
	if output == "" {
		timestamp := time.Now().Format("20060102-150405")
		dir := filepath.Dir(opts.Plan)
		statesDir := filepath.Join(dir, "dist", "states")
		if err := os.MkdirAll(statesDir, paths.DefaultDirMode); err != nil {
			return nil, err
		}
		output = filepath.Join(statesDir, fmt.Sprintf("state-%s.json", timestamp))
	}

	// Save state
	if err := newState.Write(output); err != nil {
		return nil, err
	}

	return &Result{
		State:  newState,
		Output: output,
	}, nil
}

// Loads the provider configuration.
func loadProviderConfig(providerName string) (*config.Provider, error) {
	cfg, err := config.LoadProviders()
	if err != nil {
		return nil, err
	}

	if providerName != "" {
		// Use specified provider
		return cfg.GetProvider(providerName)
	}

	// Use default provider
	provider, err := cfg.GetDefault()
	if err != nil {
		return nil, fmt.Errorf("%w: run 'crux provider add <name>' to configure a provider", config.ErrProviderNotFound)
	}
	return provider, nil
}

// Creates the appropriate deployer based on provider.
func createDeployer(provider *config.Provider) (Deployer, error) {
	switch provider.Type {
	case "aws":
		return NewAWSDeployer(provider), nil
	case "local":
		return NewLocalDeployer(provider), nil
	default:
		return nil, ErrProviderNotFound
	}
}
