package deploy

import (
	"context"

	"github.com/cruciblehq/crux/pkg/plan"
)

// Interface for deploying plans to different cloud providers.
type Deployer interface {

	// Executes a deployment plan and returns the resulting state.
	//
	// The plan 'p' contains the desired state to deploy. If currentState is
	// provided, performs an incremental deployment.
	Deploy(ctx context.Context, p *plan.Plan, currentState *plan.State) (*plan.State, error)
}
