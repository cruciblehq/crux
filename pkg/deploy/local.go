package deploy

import (
	"context"

	"github.com/cruciblehq/crux/pkg/config"
	"github.com/cruciblehq/crux/pkg/plan"
)

// Implements deployment to local Docker.
type LocalDeployer struct {
	provider config.Provider
	// TODO: Add Docker client
}

// Creates a new local deployer.
func NewLocalDeployer(provider config.Provider) *LocalDeployer {
	return &LocalDeployer{
		provider: provider,
	}
}

// Executes the deployment plan locally using Docker.
func (d *LocalDeployer) Deploy(ctx context.Context, p *plan.Plan, currentState *plan.State) (*plan.State, error) {
	// TODO: Implement local Docker deployment
	// - Start/update Docker containers for services
	// - Configure local reverse proxy
	// - Track container IDs in state
	return nil, ErrNotImplemented
}
