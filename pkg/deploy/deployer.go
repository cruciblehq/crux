package deploy

import (
	"context"

	"github.com/cruciblehq/protocol/pkg/plan"
	"github.com/cruciblehq/protocol/pkg/state"
)

// Interface for deploying plans to different cloud providers.
type Deployer interface {

	// Executes a deployment plan and returns the resulting state.
	//
	// The plan 'p' contains the desired state to deploy. If currentState is
	// provided, performs an incremental deployment.
	Deploy(ctx context.Context, p *plan.Plan, currentState *state.State) (*state.State, error)
}
