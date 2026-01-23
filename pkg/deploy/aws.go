package deploy

import (
	"context"
	"time"

	"github.com/cruciblehq/crux/pkg/config"
	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/protocol/pkg/plan"
	"github.com/cruciblehq/protocol/pkg/state"
)

// Implements deployment to AWS (ECS, ECR, ALB).
type AWSDeployer struct {
	provider *config.Provider
	// TODO: Add AWS SDK clients (ECS, ECR, ELB, etc.)
}

// Creates a new AWS deployer.
func NewAWSDeployer(provider *config.Provider) *AWSDeployer {
	return &AWSDeployer{
		provider: provider,
	}
}

// Executes a deployment plan on AWS infrastructure.
func (d *AWSDeployer) Deploy(ctx context.Context, p *plan.Plan, currentState *state.State) (*state.State, error) {
	// Calculate changes needed
	changes := d.calculateChanges(p, currentState)

	// Execute changes
	if err := d.executeChanges(ctx, changes); err != nil {
		return nil, crex.Wrap(ErrAWSOperation, err)
	}

	// Build new state from deployed resources
	newState := d.buildState(p, currentState)

	return newState, nil
}

// Determines what needs to be added, updated, or removed.
func (d *AWSDeployer) calculateChanges(p *plan.Plan, currentState *state.State) *ChangeSet {
	changes := &ChangeSet{
		ServicesToAdd:    []plan.Service{},
		ServicesToUpdate: []ServiceUpdate{},
		ServicesToRemove: []state.Service{},
	}

	// If no current state, everything is new
	if currentState == nil {
		changes.ServicesToAdd = p.Services
		return changes
	}

	// Build map of deployed services for quick lookup
	deployed := make(map[string]state.Service)
	for _, svc := range currentState.Services {
		deployed[svc.ID] = svc
	}

	// Find services to add or update
	for _, desired := range p.Services {
		if current, exists := deployed[desired.ID]; exists {
			// Service exists - check if update needed
			if d.needsUpdate(desired, current) {
				changes.ServicesToUpdate = append(changes.ServicesToUpdate, ServiceUpdate{
					Current: current,
					Desired: desired,
				})
			}
			delete(deployed, desired.ID) // Mark as handled
		} else {
			// New service
			changes.ServicesToAdd = append(changes.ServicesToAdd, desired)
		}
	}

	// Remaining deployed services need to be removed
	for _, svc := range deployed {
		changes.ServicesToRemove = append(changes.ServicesToRemove, svc)
	}

	return changes
}

// Checks if a service needs to be updated.
func (d *AWSDeployer) needsUpdate(desired plan.Service, current state.Service) bool {
	// Check if reference changed (version, digest)
	if !desired.Reference.Digest().Equal(current.Reference.Digest()) {
		return true
	}

	// TODO: Add more detailed comparison

	return false
}

// Applies the calculated changes to AWS infrastructure.
func (d *AWSDeployer) executeChanges(ctx context.Context, changes *ChangeSet) error {
	// Remove services first
	for _, svc := range changes.ServicesToRemove {
		if err := d.removeService(ctx, svc); err != nil {
			return crex.Wrap(ErrServiceOperation, err)
		}
	}

	// Add new services
	for _, svc := range changes.ServicesToAdd {
		if err := d.addService(ctx, svc); err != nil {
			return crex.Wrap(ErrServiceOperation, err)
		}
	}

	// Update existing services
	for _, update := range changes.ServicesToUpdate {
		if err := d.updateService(ctx, update); err != nil {
			return crex.Wrap(ErrServiceOperation, err)
		}
	}

	return nil
}

// Deploys a new service to AWS ECS.
func (d *AWSDeployer) addService(ctx context.Context, svc plan.Service) error {
	// TODO: Implement AWS ECS service creation
	// 1. Push image to ECR (or verify it exists)
	// 2. Create ECS task definition
	// 3. Create ECS service
	// 4. Wait for service to be healthy
	return ErrNotImplemented
}

// Updates an existing service in AWS ECS.
func (d *AWSDeployer) updateService(ctx context.Context, update ServiceUpdate) error {
	// TODO: Implement AWS ECS service update
	// 1. Update task definition with new image
	// 2. Update ECS service
	// 3. Wait for deployment to complete
	return ErrNotImplemented
}

// Removes a service from AWS ECS.
func (d *AWSDeployer) removeService(ctx context.Context, svc state.Service) error {
	// TODO: Implement AWS ECS service deletion
	// 1. Scale service to 0
	// 2. Delete ECS service
	// 3. Delete task definition (optional)
	return ErrNotImplemented
}

// Applies gateway changes to AWS ALB.
// Builds the new state from the deployed plan.
func (d *AWSDeployer) buildState(p *plan.Plan, currentState *state.State) *state.State {
	st := &state.State{
		Version: 1,
		Deployment: state.Deployment{
			DeployedAt: time.Now(),
		},
		Services: make([]state.Service, 0, len(p.Services)),
	}

	// Convert services to deployed services
	// TODO: Populate with actual AWS resource IDs
	for _, svc := range p.Services {
		deployed := state.Service{
			ID:         svc.ID,
			Reference:  svc.Reference,
			ResourceID: "arn:aws:ecs:...", // TODO: Real ARN
		}
		st.Services = append(st.Services, deployed)
	}

	return st
}

// Represents the changes needed to reach desired state.
type ChangeSet struct {
	ServicesToAdd    []plan.Service
	ServicesToUpdate []ServiceUpdate
	ServicesToRemove []state.Service
}

// Represents a service that needs updating.
type ServiceUpdate struct {
	Current state.Service
	Desired plan.Service
}
