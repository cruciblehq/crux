package resource

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
	"github.com/cruciblehq/crux/internal/resource/subsystem"
)

// [Builder] for Crucible affordances.
//
// Building an affordance resolves its grant list. References (grants with
// no Subsystem) are pulled from the registry and inlined. Domain grants are
// dispatched to the appropriate subsystem. The output is a resolved manifest
// containing only built domain grants and groups with no remaining references.
type AffordanceBuilder struct {

	// Provides registry access and namespace defaults.
	//
	// Used to pull referenced affordances from the registry during resolution.
	// The registry and default namespace determine where unqualified references
	// are resolved.
	source Source

	// Routes domain strings to subsystem implementations.
	//
	// Each entry maps a domain (e.g. "seccomp", "cap") to the subsystem that
	// handles its logic. Domains not present in this map are rejected. Each
	// subsystem knows how to build grants for its domain and apply them at
	// runtime. The builder delegates to subsystems when building domain
	// grants during resolution.
	subsystems map[subsystem.Domain]subsystem.Subsystem
}

// Returns an [AffordanceBuilder] configured with the given source.
func NewAffordanceBuilder(source Source) *AffordanceBuilder {
	return &AffordanceBuilder{
		source: source,
		subsystems: map[subsystem.Domain]subsystem.Subsystem{
			subsystem.DomainSeccomp: &subsystem.SeccompSubsystem{},
			subsystem.DomainCap:     &subsystem.CapsSubsystem{},
			subsystem.DomainFcap:    &subsystem.FcapsSubsystem{},
			subsystem.DomainCgroup:  &subsystem.CgroupSubsystem{},
		},
	}
}

// Builds a Crucible affordance resource based on the provided manifest.
//
// Affordance references in the input list are walked and expanded recursively.
// The resolved manifest written to the output directory contains only built
// domain grants with no further references.
func (ab *AffordanceBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	crex.Assertf(m.Resource.Type == manifest.TypeAffordance, "expected affordance, got %q", m.Resource.Type)

	cfg, err := manifestConfig[*manifest.Affordance](&m)
	if err != nil {
		return nil, err
	}

	resolved, err := ab.resolve(ctx, cfg.Scopes)
	if err != nil {
		return nil, err
	}

	m.Config = &manifest.Affordance{
		Schema: cfg.Schema,
		Scopes: resolved,
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

// Resolves grant scopes by expanding references and building domain grants
// via subsystems.
//
// References (grants with empty Subsystem) are pulled from the registry and
// their scopes are merged in. Domain grants are dispatched to the appropriate
// subsystem's Build method which validates, normalizes, and expands them.
func (ab *AffordanceBuilder) resolve(ctx context.Context, scopes []manifest.GrantScope) ([]manifest.GrantScope, error) {
	built := map[string][]manifest.Grant{}
	var order []string
	var extra []manifest.GrantScope

	for _, scope := range scopes {
		for _, g := range scope.Grants {
			if g.Subsystem == "" {
				ref, err := ab.resolveReference(ctx, g, scope.Platform)
				if err != nil {
					return nil, err
				}
				extra = append(extra, ref...)
				continue
			}
			var err error
			order, err = ab.buildAndCollect(ctx, g, scope.Platform, built, order)
			if err != nil {
				return nil, err
			}
		}
	}

	var result []manifest.GrantScope
	for _, p := range order {
		result = append(result, manifest.GrantScope{Platform: p, Grants: built[p]})
	}
	result = mergeScopes(result, extra)
	return result, nil
}

// Builds a domain grant and appends the results to the platform bucket.
//
// Returns the updated order slice. If the platform is seen for the first
// time, it is appended to order to preserve insertion order.
func (ab *AffordanceBuilder) buildAndCollect(ctx context.Context, g manifest.Grant, platform string, built map[string][]manifest.Grant, order []string) ([]string, error) {
	grants, err := ab.buildGrant(ctx, g)
	if err != nil {
		return order, err
	}
	if _, ok := built[platform]; !ok {
		order = append(order, platform)
	}
	built[platform] = append(built[platform], grants...)
	return order, nil
}

// Pulls a referenced affordance and returns its scopes.
func (ab *AffordanceBuilder) resolveReference(ctx context.Context, g manifest.Grant, platform string) ([]manifest.GrantScope, error) {
	if platform != "" {
		return nil, crex.Wrapf(manifest.ErrInvalidAffordance, "references cannot appear in platform-scoped grants")
	}
	aff, _, err := pullAffordance(ctx, ab.source, g.Expr)
	if err != nil {
		return nil, crex.Wrapf(ErrResolutionFailed, "pull %s: %w", g.Expr, err)
	}
	return aff.Scopes, nil
}

// Merges grant scopes, combining scopes with the same platform.
func mergeScopes(dst, src []manifest.GrantScope) []manifest.GrantScope {
	idx := map[string]int{}
	for i := range dst {
		idx[dst[i].Platform] = i
	}
	for _, s := range src {
		if i, ok := idx[s.Platform]; ok {
			dst[i].Grants = append(dst[i].Grants, s.Grants...)
		} else {
			idx[s.Platform] = len(dst)
			dst = append(dst, s)
		}
	}
	return dst
}

// Builds a domain grant by routing to the appropriate subsystem.
//
// The subsystem validates, normalizes, and expands the grant. Returns
// an error for unknown domains.
func (ab *AffordanceBuilder) buildGrant(ctx context.Context, g manifest.Grant) ([]manifest.Grant, error) {
	sub, ok := ab.subsystems[subsystem.Domain(g.Subsystem)]
	if !ok {
		return nil, crex.Wrapf(ErrResolutionFailed, "unknown domain %q", g.Subsystem)
	}
	grants, err := sub.Build(ctx, subsystem.Domain(g.Subsystem), g)
	if err != nil {
		return nil, crex.Wrapf(ErrResolutionFailed, "build %s: %w", g.Subsystem, err)
	}

	// Sanity check: each returned grant must have the same Subsystem as the input.
	for i := range grants {
		crex.Assertf(grants[i].Subsystem == g.Subsystem, "subsystem %q returned grant with subsystem %q", g.Subsystem, grants[i].Subsystem)
	}

	return grants, nil
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
