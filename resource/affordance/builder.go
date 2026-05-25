package affordance

import (
	"context"
	"strings"

	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/resource/affordance/agl"
	"github.com/cruciblehq/crux/resource/affordance/cap"
	"github.com/cruciblehq/crux/resource/affordance/cgroup"
	"github.com/cruciblehq/crux/resource/affordance/fcap"
	"github.com/cruciblehq/crux/resource/affordance/mac"
	afnet "github.com/cruciblehq/crux/resource/affordance/net"
	"github.com/cruciblehq/crux/resource/affordance/provision"
	"github.com/cruciblehq/crux/resource/affordance/rlimit"
	"github.com/cruciblehq/crux/resource/affordance/seccomp"
	"github.com/cruciblehq/crux/resource/affordance/spec"
	"github.com/cruciblehq/crux/resource/affordance/subsystem"
	"github.com/cruciblehq/crux/source"
)

// Accumulates a runtime spec from affordance grants.
//
// Accumulates state across [Builder.Build] calls. Use [NewBuilder] for each
// independent context to avoid state bleed, and [Builder.Spec] to return the
// final accumulated spec.
type Builder struct {
	spec      *spec.Spec                             // Accumulated state across all subsystems.
	provision *provision.Spec                        // Accumulated provision state.
	network   *afnet.Spec                            // Accumulated network policy state.
	subs      []subsystem.Subsystem                  // All subsystem implementations in fixed order.
	index     map[subsystem.Name]subsystem.Subsystem // Name-indexed dispatch map.
	seen      map[subsystem.Name]map[string]struct{} // Per-subsystem set of observed grant keys for conflict detection.
}

// Returns a [Builder] with all subsystems wired to a new [shared.Spec].
//
// The OCI section of the spec starts with a deny-all baseline; subsystems can
// only loosen it. Non-OCI sections start in their zero-grant state.
func NewBuilder() *Builder {
	s := spec.NewSpec()
	prov := &provision.Spec{}
	net := s.Net

	subs := []subsystem.Subsystem{
		cap.New(s.OCI.Process.Capabilities),
		rlimit.New(&s.OCI.Process.Rlimits),
		seccomp.New(s.OCI.Linux.Seccomp),
		fcap.New(s.Fcap),
		mac.New(s.MAC),
		cgroup.New(s.OCI.Linux.Resources),
		provision.New(prov),
		afnet.New(net),
	}

	idx := make(map[subsystem.Name]subsystem.Subsystem, len(subs))
	seen := make(map[subsystem.Name]map[string]struct{}, len(subs))
	for _, sub := range subs {
		idx[sub.Name()] = sub
		seen[sub.Name()] = make(map[string]struct{})
	}

	return &Builder{spec: s, provision: prov, network: net, subs: subs, index: idx, seen: seen}
}

// Returns the accumulated runtime spec produced by all processed grants.
func (b *Builder) Spec() *spec.Spec {
	return b.spec
}

// Returns the compute resource requirements accumulated from .provision grants.
func (b *Builder) Provision() *provision.Spec {
	return b.provision
}

// Returns the container network policy accumulated from .net grants.
func (b *Builder) Network() *afnet.Spec {
	return b.network
}

// Processes a single grant, updating the accumulated spec.
//
// Reference grants are pulled from the registry via source and their grants
// are processed recursively. When the reference grant carries [manifest.Grant.Args],
// they are substituted into the pulled grants before recursing. Domain grants
// are dispatched to the matching subsystem. Returns [ErrResolution] for pull
// failures, parse errors, or unknown subsystem names.
func (b *Builder) Build(ctx context.Context, g manifest.Grant, src source.Source) error {
	if g.IsRef() {
		aff, _, err := pull(ctx, src, g.RefTarget())
		if err != nil {
			return crex.Wrapf(ErrResolution, "pull %s: %w", g.RefTarget(), err)
		}
		for _, scope := range aff.Scopes {
			for _, sg := range scope.Grants {
				if err := b.Build(ctx, substituteGrant(sg, g.Args), src); err != nil {
					return err
				}
			}
		}
		return nil
	}
	parsed, err := agl.Parse(g.Source)
	if err != nil {
		return crex.Wrapf(ErrResolution, "parse %q: %w", g.Source, err)
	}
	return b.dispatch(parsed)
}

// Dispatches a parsed AGL model to the matching subsystem.
//
// Before calling Build, checks whether the subsystem's Key returns a non-empty
// string and rejects the grant with ErrConflict if an identical key has already
// been processed. Subsystems that return an empty key handle their own conflict
// detection internally.
func (b *Builder) dispatch(p *agl.Model) error {
	name := subsystem.Name(p.Subsystem)
	sub, ok := b.index[name]
	if !ok {
		return crex.Wrapf(ErrUnknownSubsystem, "unknown subsystem %q", p.Subsystem)
	}
	if key := sub.Key(p); key != "" {
		if _, dup := b.seen[name][key]; dup {
			return crex.Wrapf(ErrConflict, "duplicate %s grant %q", p.Subsystem, key)
		}
		if err := sub.Build(p); err != nil {
			return err
		}
		b.seen[name][key] = struct{}{}
		return nil
	}
	return sub.Build(p)
}

// Resolves grant args into a substitution map keyed by parameter name.
//
// When args is non-empty it is returned as-is. When value is set the
// affordance schema must declare a default parameter; the returned map
// contains a single entry mapping the default parameter name to value.
// Returns nil when both are zero.
// Returns a copy of g with all $param references in Source replaced by the
// corresponding value from params.
//
// Substitution is plain string replacement. Params with names that do not
// appear in Source are silently ignored. Returns g unchanged when params is nil.
func substituteGrant(g manifest.Grant, params map[string]string) manifest.Grant {
	if len(params) == 0 {
		return g
	}
	src := g.Source
	for k, v := range params {
		src = strings.ReplaceAll(src, "$"+k, v)
	}
	return manifest.Grant{Source: src}
}

// Fetches an affordance resource and returns its config and content digest.
//
// Uses [Source.Pull) to fetch the resource
func pull(ctx context.Context, src source.Source, target string) (*manifest.Affordance, string, error) {
	ref, err := src.Parse(string(manifest.TypeAffordance), target)
	if err != nil {
		return nil, "", err
	}
	result, err := src.Pull(ctx, ref)
	if err != nil {
		return nil, "", err
	}
	aff, err := manifest.ReadAsAt[*manifest.Affordance](result.Extracted)
	if err != nil {
		return nil, "", crex.Wrapf(ErrResolution, "%s: %v", target, err)
	}
	return aff, result.Digest, nil
}
