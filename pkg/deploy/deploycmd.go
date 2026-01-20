package deploy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cruciblehq/crux/pkg/config"
	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/paths"
	"github.com/cruciblehq/crux/pkg/plan"
)

var (
	ErrInvalidPlan      = errors.New("invalid plan file")
	ErrInvalidState     = errors.New("invalid state file")
	ErrProviderNotFound = errors.New("provider not found")
)

// Options for executing a deployment.
type DeployOptions struct {
	PlanPath     string // Path to plan file
	StatePath    string // Optional path to existing state for incremental deployment
	OutputPath   string // Optional output path for state file
	ProviderName string // Optional provider configuration name (uses default if empty)
}

// Result of a deployment execution.
type DeployResult struct {
	State      *plan.State // New state after deployment
	OutputPath string      // Path where state was saved
}

// Execute executes a deployment plan.
func Execute(ctx context.Context, opts DeployOptions) (*DeployResult, error) {
	// Load provider configuration
	providerConfig, err := loadProviderConfig(opts.ProviderName)
	if err != nil {
		return nil, err
	}

	// Read plan
	p, err := plan.Read(opts.PlanPath)
	if err != nil {
		return nil, crex.Wrap(ErrInvalidPlan, err)
	}

	// Read existing state if provided
	var currentState *plan.State
	if opts.StatePath != "" {
		currentState, err = plan.ReadState(opts.StatePath)
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
	outputPath := opts.OutputPath
	if outputPath == "" {
		timestamp := time.Now().Format("20060102-150405")
		dir := filepath.Dir(opts.PlanPath)
		statesDir := filepath.Join(dir, "dist", "states")
		if err := os.MkdirAll(statesDir, paths.DefaultDirMode); err != nil {
			return nil, err
		}
		outputPath = filepath.Join(statesDir, fmt.Sprintf("state-%s.json", timestamp))
	}

	// Save state
	if err := newState.Write(outputPath); err != nil {
		return nil, err
	}

	return &DeployResult{
		State:      newState,
		OutputPath: outputPath,
	}, nil
}

// Loads the provider configuration.
func loadProviderConfig(providerName string) (config.Provider, error) {
	cfg, err := config.LoadProviders()
	if err != nil {
		return config.Provider{}, err
	}

	if providerName != "" {
		// Use specified provider
		return cfg.GetProvider(providerName)
	}

	// Use default provider
	provider, err := cfg.GetDefault()
	if err != nil {
		return config.Provider{}, fmt.Errorf("%w: run 'crux provider add <name>' to configure a provider", config.ErrProviderNotFound)
	}
	return provider, nil
}

// Creates the appropriate deployer based on provider.
func createDeployer(provider config.Provider) (Deployer, error) {
	switch provider.Type {
	case "aws":
		return NewAWSDeployer(provider), nil
	case "local":
		return NewLocalDeployer(provider), nil
	default:
		return nil, ErrProviderNotFound
	}
}
