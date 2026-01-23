package plan

import (
	"context"
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
func resolveServiceReferences(ctx context.Context, bp *blueprint.Blueprint, st *state.State, registryClient *registry.Client, p *plan.Plan) error {
	for i := range bp.Services {
		service := &bp.Services[i]

		ref, err := reference.Parse(service.Reference, resource.TypeService, nil)
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

		addServiceToPlan(service, frozenRef, p)
	}

	return nil
}

// Adds a service to the plan.
//
// Creates a service entry with the frozen reference and adds it to the plan
// along with a corresponding gateway route. The route maps the service's prefix
// path to the service instance, allowing the gateway to forward requests.
func addServiceToPlan(svc *blueprint.Service, frozenRef *reference.Reference, p *plan.Plan) {
	service := plan.Service{
		ID:        svc.ID,
		Reference: *frozenRef,
	}
	p.Services = append(p.Services, service)

	route := plan.Route{
		Pattern:   svc.Prefix,
		ServiceID: svc.ID,
	}
	p.Gateway.Routes = append(p.Gateway.Routes, route)
}

// Resolves a reference to a frozen reference with digest.
//
// Takes a reference from the blueprint (which may use version constraints or
// channels) and resolves it against the registry to determine the exact version
// with digest. Returns an error if the reference cannot be resolved.
func resolveReference(ctx context.Context, client *registry.Client, ref *reference.Reference) (*reference.Reference, error) {
	namespace := ref.Namespace()
	resourceName := ref.Name()

	selectedVersion, err := resolveVersion(ctx, client, ref, namespace, resourceName)
	if err != nil {
		return nil, err
	}

	if selectedVersion.Digest == nil {
		return nil, ErrMissingDigest
	}

	return buildFrozenReference(ref, namespace, resourceName, selectedVersion)
}

// Resolves a reference to a specific version.
//
// For channel-based references, follows the channel to get the version. For
// version-constrained references, finds the latest version matching the constraint.
func resolveVersion(ctx context.Context, client *registry.Client, ref *reference.Reference, namespace, resourceName string) (*registry.Version, error) {
	if ref.IsChannelBased() {
		channelName := *ref.Channel()
		channel, err := client.ReadChannel(ctx, namespace, resourceName, channelName)
		if err != nil {
			return nil, crex.Wrap(ErrChannelNotFound, err)
		}
		return &channel.Version, nil
	}

	versions, err := client.ListVersions(ctx, namespace, resourceName)
	if err != nil {
		return nil, crex.Wrap(ErrCannotListVersions, err)
	}

	if len(versions.Versions) == 0 {
		return nil, ErrNoMatchingVersion
	}

	latestVersion, err := findLatestVersion(versions.Versions, ref.Version())
	if err != nil {
		return nil, err
	}

	selectedVersion, err := client.ReadVersion(ctx, namespace, resourceName, latestVersion.String())
	if err != nil {
		return nil, crex.Wrap(ErrCannotReadVersion, err)
	}

	return selectedVersion, nil
}

// Builds a frozen reference from a resolved version.
//
// Constructs a frozen reference directly from components, avoiding the overhead
// of string formatting and parsing.
func buildFrozenReference(ref *reference.Reference, namespace, resourceName string, version *registry.Version) (*reference.Reference, error) {
	id := reference.NewIdentifier(ref.Type(), namespace, resourceName)

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

// Finds the latest version matching the constraint.
//
// Iterates through all versions, filtering by the constraint and comparing
// semantic versions to find the highest match. Returns the parsed version of
// the latest match or an error if no versions satisfy the constraint.
func findLatestVersion(versions []registry.VersionSummary, constraint *reference.VersionConstraint) (*reference.Version, error) {
	var latestVersion *reference.Version

	for _, v := range versions {
		parsedVersion := tryParseMatchingVersion(v, constraint)
		if parsedVersion == nil {
			continue
		}

		if latestVersion == nil {
			latestVersion = parsedVersion
		} else {
			cmp, valid := parsedVersion.Compare(latestVersion)
			if valid && cmp > 0 {
				latestVersion = parsedVersion
			}
		}
	}

	if latestVersion == nil {
		return nil, ErrNoMatchingVersion
	}

	return latestVersion, nil
}

// Attempts to parse and validate a version against the given constraint.
//
// Returns the parsed version if it matches the constraint, or nil if parsing
// failed, the version doesn't match, or any other error occurred.
func tryParseMatchingVersion(v registry.VersionSummary, constraint *reference.VersionConstraint) *reference.Version {
	parsedVersion, err := reference.ParseVersion(v.String)
	if err != nil {
		slog.Warn("skipping malformed version from registry", "version", v.String, "error", err.Error())
		return nil
	}

	matches, err := constraint.MatchesVersion(parsedVersion)
	if err != nil {
		slog.Warn("skipping malformed version from registry", "version", v.String, "error", err.Error())
		return nil
	}

	if !matches {
		return nil
	}

	return parsedVersion
}
