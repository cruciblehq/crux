package resource

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
)

// [Builder] for Crucible affordances.
//
// Building an affordance resolves references in its grant list. Grants with
// subsystem "ref" are pulled from the registry and expanded recursively.
// Domain grants and platform groups are preserved as-is. The output is a
// resolved manifest whose grant list contains only domain grants and groups
// with no remaining references.
type AffordanceBuilder struct {
	source Source // Registry and namespace defaults for pulling child affordances.
}

// Returns an [AffordanceBuilder] configured with the given source.
func NewAffordanceBuilder(source Source) *AffordanceBuilder {
	return &AffordanceBuilder{source: source}
}

// Builds a Crucible affordance resource based on the provided manifest.
//
// Affordance references in the grant list are walked and expanded recursively.
// The resolved manifest written to the output directory contains only domain
// grants with no further references.
func (ab *AffordanceBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	cfg, err := manifestConfig[*manifest.Affordance](&m)
	if err != nil {
		return nil, err
	}

	resolved, err := ab.resolve(ctx, cfg.Grants)
	if err != nil {
		return nil, err
	}

	m.Config = &manifest.Affordance{
		Schema: cfg.Schema,
		Grants: resolved,
	}

	if err := WriteManifest(&m, output); err != nil {
		return nil, err
	}

	return &BuildResult{Output: output, Manifest: &m}, nil
}

// Verifies that the build directory contains the expected affordance artifacts.
func (ab *AffordanceBuilder) Verify(buildDir string) error {
	return verify(buildDir, manifest.TypeAffordance, "")
}

// Packages the affordance's build output into a distributable archive.
func (ab *AffordanceBuilder) Pack(ctx context.Context, buildDir, output string) (*PackResult, error) {
	return pack(ctx, buildDir, output)
}

// Uploads an affordance package archive to the Hub registry.
func (ab *AffordanceBuilder) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return push(ctx, ab.source.Registry, m, packagePath)
}

// Resolves grants by expanding affordance references.
//
// References (subsystem "ref") are pulled from the registry and expanded
// recursively. Domain grants and platform groups are collected as-is. Groups
// are walked to resolve any refs within their children.
func (ab *AffordanceBuilder) resolve(ctx context.Context, grants []manifest.Grant) ([]manifest.Grant, error) {
	var result []manifest.Grant
	visited := make(map[string]bool)
	if err := ab.walk(ctx, grants, &result, visited); err != nil {
		return nil, err
	}
	return result, nil
}

// Resolves a list of affordance refs into grants.
//
// Each ref targets an affordance resource. The affordance is pulled from the
// registry and its grants are resolved recursively. Used by the blueprint
// builder to resolve service-level affordance refs.
func (ab *AffordanceBuilder) resolveRefs(ctx context.Context, refs []manifest.Ref) ([]manifest.Grant, error) {
	var result []manifest.Grant
	visited := make(map[string]bool)

	for _, ref := range refs {
		aff, digest, err := pullAffordance(ctx, ab.source, ref.Target)
		if err != nil {
			return nil, crex.Wrapf(ErrResolutionFailed, "pull %s: %w", ref.Target, err)
		}

		if visited[digest] {
			return nil, crex.Wrapf(ErrResolutionCycle, "%s", ref.Target)
		}
		visited[digest] = true

		if err := ab.walk(ctx, aff.Grants, &result, visited); err != nil {
			return nil, err
		}

		delete(visited, digest)
	}

	return result, nil
}

// Recursively resolves grants.
//
// Domain grants are collected as-is. Platform groups are walked to resolve
// refs within their children, then collected with the resolved children.
// Reference grants (subsystem "ref") are pulled from the registry and
// their grants are walked recursively. Cycles are detected by content digest.
func (ab *AffordanceBuilder) walk(ctx context.Context, grants []manifest.Grant, result *[]manifest.Grant, visited map[string]bool) error {
	for _, grant := range grants {
		// Platform group: recurse into children, rebuild group.
		if len(grant.Grants) > 0 {
			var children []manifest.Grant
			if err := ab.walk(ctx, grant.Grants, &children, visited); err != nil {
				return err
			}
			*result = append(*result, manifest.Grant{
				Platform: grant.Platform,
				Grants:   children,
			})
			continue
		}

		// Domain grant: keep as-is.
		if grant.Subsystem != manifest.SubRef {
			*result = append(*result, grant)
			continue
		}

		// Reference: resolve and recurse.
		aff, digest, err := pullAffordance(ctx, ab.source, grant.Expr)
		if err != nil {
			return crex.Wrapf(ErrResolutionFailed, "pull %s: %w", grant.Expr, err)
		}

		if visited[digest] {
			return crex.Wrapf(ErrResolutionCycle, "%s", grant.Expr)
		}
		visited[digest] = true

		if err := ab.walk(ctx, aff.Grants, result, visited); err != nil {
			return err
		}

		delete(visited, digest)
	}
	return nil
}

// Pulls an affordance resource and extracts its configuration.
//
// Returns the affordance config and its content digest.
func pullAffordance(ctx context.Context, source Source, target string) (*manifest.Affordance, string, error) {
	ref, err := source.Parse(manifest.TypeAffordance, target)
	if err != nil {
		return nil, "", err
	}

	result, err := source.Pull(ctx, ref)
	if err != nil {
		return nil, "", err
	}

	m, err := ReadManifestIn(result.Dir)
	if err != nil {
		return nil, "", err
	}

	aff, ok := m.Config.(*manifest.Affordance)
	if !ok {
		return nil, "", crex.Wrapf(ErrResolutionFailed, "%s is not an affordance", target)
	}
	return aff, result.Digest, nil
}
