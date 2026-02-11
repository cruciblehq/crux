package blueprint

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/cruciblehq/crux/config"
	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/plan"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/registry"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/state"
)

const (

	// Default compute ID (placeholder).
	defaultComputeID = "main-compute"

	// Default compute instance type for all AWS deployments.
	defaultAWSInstanceType = "t3.micro"
)

// Options for generating a deployment plan from a blueprint.
type ExecuteOptions struct {
	State            string              // Path to existing state for incremental planning (optional).
	Registry         string              // Registry URL for resolving references.
	Provider         config.ProviderType // Provider type.
	DefaultNamespace string              // Default namespace for resource identifiers.
}

// Generates a deployment plan from the blueprint.
//
// Resolves all service references against the registry, allocates compute
// resources, creates bindings, and configures gateway routing. If state is
// provided, enables incremental planning based on current deployment state.
func (bp *Blueprint) Execute(ctx context.Context, opts ExecuteOptions) (*plan.Plan, error) {
	// Read existing state if provided
	var st *state.State
	if opts.State != "" {
		var err error
		st, err = state.Read(opts.State)
		if err != nil {
			return nil, crex.UserError("invalid state file", err.Error()).
				Fallback("Ensure the state file exists and is valid.").
				Err()
		}
	}

	p := &plan.Plan{
		Version:      plan.Version,
		Services:     make([]plan.Service, 0, len(bp.Services)),
		Compute:      make([]plan.Compute, 0, 1),
		Environments: make([]plan.Environment, 0),
		Bindings:     make([]plan.Binding, 0, len(bp.Services)),
		Gateway: plan.Gateway{
			Routes: make([]plan.Route, 0, len(bp.Services)),
		},
	}

	registryClient := registry.NewClient(opts.Registry, nil)

	// Resolve all service references
	if err := resolveServiceReferences(ctx, bp, st, registryClient, opts.Registry, opts.DefaultNamespace, p); err != nil {
		return nil, err
	}

	// Allocate compute resources
	allocateCompute(p, opts.Provider, string(opts.Provider))

	// Create bindings between services and compute
	bind(p)

	return p, nil
}

// Resolves all service references in the blueprint and adds them to the plan.
func resolveServiceReferences(ctx context.Context, bp *Blueprint, st *state.State, registryClient *registry.Client, registryHost string, defaultNamespace string, p *plan.Plan) error {
	slog.Info("resolving service references", "registryHost", registryHost, "serviceCount", len(bp.Services))

	if st != nil {
		slog.Warn("service reference resolution does not currently support incremental planning; ignoring existing state")
	}

	for i := range bp.Services {
		service := &bp.Services[i]

		opts, err := reference.NewIdentifierOptions(registryHost, defaultNamespace)
		if err != nil {
			return err
		}

		ref, err := reference.Parse(service.Reference, resource.TypeService, opts)
		if err != nil {
			return crex.UserError("invalid service reference in blueprint", fmt.Sprintf("cannot parse reference '%s'", service.Reference)).
				Fallback("Ensure the service reference follows the expected format.").
				Cause(err).
				Err()
		}

		frozenRef, err := resolveReference(ctx, registryClient, ref)
		if err != nil {
			return crex.UserError(fmt.Sprintf("cannot resolve service reference %s", service.Reference), err.Error()).
				Fallback("Ensure the service exists in the registry and the version constraint is satisfiable.").
				Err()
		}

		if err := addServiceToPlan(service, frozenRef, p); err != nil {
			return err
		}
	}

	return nil
}

// Resolves a reference to a frozen reference with digest.
func resolveReference(ctx context.Context, client *registry.Client, ref *reference.Reference) (*reference.Reference, error) {
	selectedVersion, err := registry.ResolveVersion(ctx, client, ref)
	if err != nil {
		return nil, wrapResolveError(err)
	}

	if selectedVersion.Digest == nil {
		return nil, ErrMissingDigest
	}

	return buildFrozenReference(ref, ref.Namespace(), ref.Name(), selectedVersion)
}

// Wraps resolution errors with plan-specific sentinel errors for context.
func wrapResolveError(err error) error {
	if errors.Is(err, registry.ErrNoVersions) || errors.Is(err, registry.ErrNoMatchingVersion) {
		return crex.Wrap(ErrNoMatchingVersion, err)
	}

	var regErr *registry.Error
	if errors.As(err, &regErr) && regErr.Code == registry.ErrorCodeNotFound {
		return crex.Wrap(ErrChannelNotFound, err)
	}

	return crex.Wrap(ErrCannotReadVersion, err)
}

// Builds a frozen reference from a resolved version.
func buildFrozenReference(ref *reference.Reference, namespace, resourceName string, version *registry.Version) (*reference.Reference, error) {
	reg := ref.Registry()
	id, err := reference.NewIdentifier(
		ref.Type(),
		reg.String(),
		namespace,
		resourceName,
	)
	if err != nil {
		return nil, crex.Wrap(ErrCannotCreateReference, err)
	}

	digest, err := reference.ParseDigest(*version.Digest)
	if err != nil {
		return nil, crex.Wrap(ErrCannotCreateReference, err)
	}

	frozenRef, err := reference.New(id, version.String, digest)
	if err != nil {
		return nil, crex.Wrap(ErrCannotCreateReference, err)
	}

	return frozenRef, nil
}

// Adds a service to the plan.
func addServiceToPlan(svc *Service, frozenRef *reference.Reference, p *plan.Plan) error {
	if svc.ID == "" {
		return crex.UserError("invalid service in blueprint", "service ID is required").
			Fallback("Add an 'id' field to each service in the blueprint.").
			Err()
	}

	service := plan.Service{
		ID:        svc.ID,
		Reference: frozenRef.String(),
	}
	p.Services = append(p.Services, service)

	route := plan.Route{
		Pattern: svc.Prefix,
		Service: svc.ID,
	}
	p.Gateway.Routes = append(p.Gateway.Routes, route)
	return nil
}

// Allocates compute resources for the deployment plan.
func allocateCompute(p *plan.Plan, providerType config.ProviderType, providerName string) {
	compute := plan.Compute{
		ID:       defaultComputeID,
		Provider: providerName,
		Config:   computeConfigForProvider(providerType),
	}
	p.Compute = append(p.Compute, compute)
}

// Returns provider-specific compute configuration.
func computeConfigForProvider(providerType config.ProviderType) any {
	switch providerType {
	case config.ProviderTypeAWS:
		return plan.ComputeAWS{
			InstanceType: defaultAWSInstanceType,
		}
	case config.ProviderTypeLocal:
		return nil
	default:
		return nil
	}
}

// Creates bindings between services and compute resources.
func bind(p *plan.Plan) {
	computeID := defaultComputeID
	for _, service := range p.Services {
		binding := plan.Binding{
			Service: service.ID,
			Compute: computeID,
		}
		p.Bindings = append(p.Bindings, binding)
	}
}
