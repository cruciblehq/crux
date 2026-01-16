package deploy

import (
	"context"
	"errors"
	"time"

	"github.com/cruciblehq/crux/pkg/config"
	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/plan"
)

var (
	ErrAWSOperation     = errors.New("aws operation failed")
	ErrNotImplemented   = errors.New("not implemented")
	ErrServiceOperation = errors.New("service operation failed")
	ErrGatewayOperation = errors.New("gateway operation failed")
)

// AWSDeployer implements deployment to AWS (ECS, ECR, ALB).
type AWSDeployer struct {
	provider config.Provider
	// TODO: Add AWS SDK clients (ECS, ECR, ELB, etc.)
}

// NewAWSDeployer creates a new AWS deployer.
func NewAWSDeployer(provider config.Provider) *AWSDeployer {
	return &AWSDeployer{
		provider: provider,
	}
}

// Deploy executes a deployment plan on AWS infrastructure.
func (d *AWSDeployer) Deploy(ctx context.Context, p *plan.Plan, currentState *plan.State) (*plan.State, error) {
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

// calculateChanges determines what needs to be added, updated, or removed.
func (d *AWSDeployer) calculateChanges(p *plan.Plan, currentState *plan.State) *ChangeSet {
	changes := &ChangeSet{
		ServicesToAdd:    []plan.ResolvedService{},
		ServicesToUpdate: []ServiceUpdate{},
		ServicesToRemove: []plan.DeployedService{},
		GatewayChanges:   &GatewayChange{},
	}

	// If no current state, everything is new
	if currentState == nil {
		changes.ServicesToAdd = p.Services
		if p.Gateway != nil {
			changes.GatewayChanges.Action = "create"
			changes.GatewayChanges.NewGateway = p.Gateway
		}
		return changes
	}

	// Build map of deployed services for quick lookup
	deployed := make(map[string]plan.DeployedService)
	for _, svc := range currentState.Services {
		deployed[svc.Reference] = svc
	}

	// Find services to add or update
	for _, desired := range p.Services {
		if current, exists := deployed[desired.Reference]; exists {
			// Service exists - check if update needed
			if d.needsUpdate(desired, current) {
				changes.ServicesToUpdate = append(changes.ServicesToUpdate, ServiceUpdate{
					Current: current,
					Desired: desired,
				})
			}
			delete(deployed, desired.Reference) // Mark as handled
		} else {
			// New service
			changes.ServicesToAdd = append(changes.ServicesToAdd, desired)
		}
	}

	// Remaining deployed services need to be removed
	for _, svc := range deployed {
		changes.ServicesToRemove = append(changes.ServicesToRemove, svc)
	}

	// Check gateway changes
	if p.Gateway != nil && currentState.Gateway != nil {
		if d.gatewayNeedsUpdate(p.Gateway, currentState.Gateway) {
			changes.GatewayChanges.Action = "update"
			changes.GatewayChanges.Current = currentState.Gateway
			changes.GatewayChanges.NewGateway = p.Gateway
		}
	} else if p.Gateway != nil && currentState.Gateway == nil {
		changes.GatewayChanges.Action = "create"
		changes.GatewayChanges.NewGateway = p.Gateway
	} else if p.Gateway == nil && currentState.Gateway != nil {
		changes.GatewayChanges.Action = "delete"
		changes.GatewayChanges.Current = currentState.Gateway
	}

	return changes
}

// needsUpdate checks if a service needs to be updated.
func (d *AWSDeployer) needsUpdate(desired plan.ResolvedService, current plan.DeployedService) bool {
	// Check if image digest changed
	if desired.ImageDigest != current.ImageDigest {
		return true
	}

	// Check if configuration changed
	// TODO: Add more detailed comparison

	return false
}

// gatewayNeedsUpdate checks if gateway configuration changed.
func (d *AWSDeployer) gatewayNeedsUpdate(desired *plan.Gateway, current *plan.DeployedGateway) bool {
	// Check if routes changed
	if len(desired.Routes) != len(current.Routes) {
		return true
	}

	// TODO: Add more detailed comparison

	return false
}

// executeChanges applies the calculated changes to AWS infrastructure.
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

	// Handle gateway changes
	if changes.GatewayChanges.Action != "" {
		if err := d.applyGatewayChanges(ctx, changes.GatewayChanges); err != nil {
			return crex.Wrap(ErrGatewayOperation, err)
		}
	}

	return nil
}

// addService deploys a new service to AWS ECS.
func (d *AWSDeployer) addService(ctx context.Context, svc plan.ResolvedService) error {
	// TODO: Implement AWS ECS service creation
	// 1. Push image to ECR (or verify it exists)
	// 2. Create ECS task definition
	// 3. Create ECS service
	// 4. Wait for service to be healthy
	return ErrNotImplemented
}

// updateService updates an existing service in AWS ECS.
func (d *AWSDeployer) updateService(ctx context.Context, update ServiceUpdate) error {
	// TODO: Implement AWS ECS service update
	// 1. Update task definition with new image
	// 2. Update ECS service
	// 3. Wait for deployment to complete
	return ErrNotImplemented
}

// removeService removes a service from AWS ECS.
func (d *AWSDeployer) removeService(ctx context.Context, svc plan.DeployedService) error {
	// TODO: Implement AWS ECS service deletion
	// 1. Scale service to 0
	// 2. Delete ECS service
	// 3. Delete task definition (optional)
	return ErrNotImplemented
}

// applyGatewayChanges creates, updates, or deletes the gateway (ALB).
func (d *AWSDeployer) applyGatewayChanges(ctx context.Context, changes *GatewayChange) error {
	// TODO: Implement AWS ALB management
	// 1. Create/update ALB
	// 2. Configure target groups
	// 3. Configure listener rules
	return ErrNotImplemented
}

// buildState constructs the new state from the deployed plan.
func (d *AWSDeployer) buildState(p *plan.Plan, currentState *plan.State) *plan.State {
	state := &plan.State{
		Plan:       "", // TODO: Store plan reference/digest
		DeployedAt: time.Now(),
		Services:   make([]plan.DeployedService, 0, len(p.Services)),
	}

	// Convert resolved services to deployed services
	// TODO: Populate with actual AWS resource IDs and metadata
	for _, svc := range p.Services {
		deployed := plan.DeployedService{
			Reference:    svc.Reference,
			Resolved:     svc.Resolved,
			Prefix:       svc.Prefix,
			ImageDigest:  svc.ImageDigest,
			Provider:     "aws",
			ResourceType: "ecs-service",
			ResourceID:   "arn:aws:ecs:...", // TODO: Real ARN
			Status:       "running",
			Metadata:     make(map[string]string),
		}
		state.Services = append(state.Services, deployed)
	}

	// Convert gateway if present
	if p.Gateway != nil {
		state.Gateway = &plan.DeployedGateway{
			Type:         p.Gateway.Type,
			Listen:       p.Gateway.Listen,
			Provider:     "aws",
			ResourceType: "alb",
			ResourceID:   "arn:aws:elasticloadbalancing:...", // TODO: Real ARN
			Status:       "active",
			Routes:       make([]plan.DeployedRoute, 0, len(p.Gateway.Routes)),
			Metadata:     make(map[string]string),
		}

		for _, route := range p.Gateway.Routes {
			state.Gateway.Routes = append(state.Gateway.Routes, plan.DeployedRoute{
				Prefix:   route.Prefix,
				Upstream: route.Upstream,
				Service:  route.Service,
				RuleID:   "rule-...", // TODO: Real rule ID
			})
		}
	}

	return state
}

// ChangeSet represents the changes needed to reach desired state.
type ChangeSet struct {
	ServicesToAdd    []plan.ResolvedService
	ServicesToUpdate []ServiceUpdate
	ServicesToRemove []plan.DeployedService
	GatewayChanges   *GatewayChange
}

// ServiceUpdate represents a service that needs updating.
type ServiceUpdate struct {
	Current plan.DeployedService
	Desired plan.ResolvedService
}

// GatewayChange represents gateway configuration changes.
type GatewayChange struct {
	Action     string // create, update, delete
	Current    *plan.DeployedGateway
	NewGateway *plan.Gateway
}
