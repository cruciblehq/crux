package deploy

import (
	"context"

	"github.com/cruciblehq/crux/pkg/config"
	"github.com/cruciblehq/protocol/pkg/plan"
	"github.com/cruciblehq/protocol/pkg/state"
)

// Implements deployment to local Docker.
type LocalDeployer struct {
	provider *config.Provider
	// TODO: Add Docker client
}

// Creates a new local deployer.
func NewLocalDeployer(provider *config.Provider) *LocalDeployer {
	return &LocalDeployer{
		provider: provider,
	}
}

// Executes the deployment plan locally using Docker.
func (d *LocalDeployer) Deploy(ctx context.Context, p *plan.Plan, currentState *state.State) (*state.State, error) {
	// TODO: Implement local Docker deployment
	// - Start/update Docker containers for services
	// - Configure local reverse proxy
	// - Track container IDs in state
	return nil, ErrNotImplemented
}
