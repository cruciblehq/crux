package plan

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/protocol/pkg/blueprint"
	"github.com/cruciblehq/protocol/pkg/plan"
	"github.com/cruciblehq/protocol/pkg/reference"
	"github.com/cruciblehq/protocol/pkg/registry"
	"github.com/cruciblehq/protocol/pkg/resource"
	"github.com/cruciblehq/protocol/pkg/state"
)

// Resolves all service references in the blueprint and adds them to the plan.
//
// Each service is resolved independently to a frozen reference (with digest).
// The resolution queries the registry to find the latest matching version for
// version constraints or follows channel pointers to specific versions.
func resolveServiceReferences(ctx context.Context, bp *blueprint.Blueprint, st *state.State, registryClient *registry.Client, registryHost string, p *plan.Plan) error {
	slog.Info("resolving service references", "registryHost", registryHost, "serviceCount", len(bp.Services))

	if st != nil {
		slog.Warn("service reference resolution does not currently support incremental planning; ignoring existing state")
	}

	for i := range bp.Services {
		service := &bp.Services[i]

		opts := &reference.IdentifierOptions{
			DefaultRegistry: registryHost,
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
//
// Takes a reference from the blueprint (which may use version constraints or
// channels) and resolves it against the registry to determine the exact version
// with digest. Returns an error if the reference cannot be resolved.
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
	// Check for local resolution errors (from ResolveVersion)
	if errors.Is(err, registry.ErrNoVersions) || errors.Is(err, registry.ErrNoMatchingVersion) {
		return crex.Wrap(ErrNoMatchingVersion, err)
	}

	// Check for API errors (channel not found, etc.)
	var regErr *registry.Error
	if errors.As(err, &regErr) && regErr.Code == registry.ErrorCodeNotFound {
		return crex.Wrap(ErrChannelNotFound, err)
	}

	return crex.Wrap(ErrCannotReadVersion, err)
}

// Builds a frozen reference from a resolved version.
//
// Constructs a frozen reference directly from components, avoiding the overhead
// of string formatting and parsing.
func buildFrozenReference(ref *reference.Reference, namespace, resourceName string, version *registry.Version) (*reference.Reference, error) {
	id := reference.NewIdentifier(
		ref.Type(),
		ref.Registry(),
		namespace,
		resourceName,
	)

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
//
// Creates a service entry with the frozen reference and adds it to the plan
// along with a corresponding gateway route. The route maps the service's prefix
// path to the service instance, allowing the gateway to forward requests.
func addServiceToPlan(svc *blueprint.Service, frozenRef *reference.Reference, p *plan.Plan) error {
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

	// Service routes are added to the gateway
	p.Gateway.Routes = append(p.Gateway.Routes, route)
	return nil
}
