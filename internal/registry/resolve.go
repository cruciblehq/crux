package registry

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/reference"
)

// Resolves a reference to a specific version with full metadata.
//
// For channel-based references, follows the channel to get the version. For
// version-constrained references, lists all versions and finds the highest
// matching version. Returns the full Version with digest and archive details.
func ResolveVersion(ctx context.Context, client *Client, ref *reference.Reference) (*Version, error) {
	if err := validateResourceType(ctx, client, ref); err != nil {
		return nil, err
	}

	if ref.IsChannelBased() {
		return resolveChannel(ctx, client, ref)
	}
	return resolveVersionConstraint(ctx, client, ref)
}

// Validates that the resource type matches the reference type.
func validateResourceType(ctx context.Context, client *Client, ref *reference.Reference) error {
	res, err := client.ReadResource(ctx, ref.Namespace(), ref.Name())
	if err != nil {
		return err
	}

	if res.Type != string(ref.Type()) {
		return ErrTypeMismatch
	}
	return nil
}

// Resolves a channel reference to its current version.
func resolveChannel(ctx context.Context, client *Client, ref *reference.Reference) (*Version, error) {
	channel, err := client.ReadChannel(ctx, ref.Namespace(), ref.Name(), *ref.Channel())
	if err != nil {
		return nil, err
	}
	return &channel.Version, nil
}

// Resolves a version constraint to the highest matching version.
func resolveVersionConstraint(ctx context.Context, client *Client, ref *reference.Reference) (*Version, error) {
	versions, err := client.ListVersions(ctx, ref.Namespace(), ref.Name())
	if err != nil {
		return nil, err
	}

	if len(versions.Versions) == 0 {
		return nil, crex.Wrapf(ErrNoVersions, "%s/%s", ref.Namespace(), ref.Name())
	}

	latestVersion := FindLatestVersion(versions.Versions, ref.Version())
	if latestVersion == nil {
		return nil, crex.Wrapf(ErrNoMatchingVersion, "%s", ref.Version())
	}

	return client.ReadVersion(ctx, ref.Namespace(), ref.Name(), latestVersion.String())
}

// Finds the highest version from a list that satisfies the given constraint.
//
// Iterates through all versions, filtering by the constraint and comparing
// semantic versions to find the highest match. Returns the parsed version of
// the latest match or nil if no versions satisfy the constraint.
func FindLatestVersion(versions []VersionSummary, constraint *reference.VersionConstraint) *reference.Version {
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

	return latestVersion
}

// Attempts to parse and validate a version against the given constraint.
//
// Returns the parsed version if it matches the constraint, or nil if parsing
// failed, the version doesn't match, or any other error occurred.
func tryParseMatchingVersion(v VersionSummary, constraint *reference.VersionConstraint) *reference.Version {
	parsedVersion, err := reference.ParseVersion(v.String)
	if err != nil {
		slog.Warn("skipping malformed version from registry", "version", v.String, "error", err.Error())
		return nil
	}

	matches, err := constraint.MatchesVersion(parsedVersion)
	if err != nil {
		slog.Warn("error checking version constraint", "version", v.String, "error", err.Error())
		return nil
	}

	if !matches {
		return nil
	}

	return parsedVersion
}
