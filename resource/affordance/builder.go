package affordance

import (
	"context"
	"strings"

	"github.com/cruciblehq/crux/hub"
	aff "github.com/cruciblehq/spec/affordance"
	"github.com/cruciblehq/spec/affordance/agl"
	"github.com/cruciblehq/spec/affordance/cap"
	"github.com/cruciblehq/spec/affordance/cgroup"
	"github.com/cruciblehq/spec/affordance/device"
	"github.com/cruciblehq/spec/affordance/fcap"
	"github.com/cruciblehq/spec/affordance/kernel"
	"github.com/cruciblehq/spec/affordance/mac"
	"github.com/cruciblehq/spec/affordance/mount"
	"github.com/cruciblehq/spec/affordance/net"
	"github.com/cruciblehq/spec/affordance/provision"
	"github.com/cruciblehq/spec/affordance/rlimit"
	"github.com/cruciblehq/spec/affordance/seccomp"
	"github.com/cruciblehq/spec/affordance/subsystem"
	"github.com/cruciblehq/spec/affordance/volume"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/crex"
)

// Compiles an [aff.Spec] from affordance grants.
//
// Accumulates state for [Builder.Build]. Use [NewBuilder] for each build to
// avoid state bleed, and [Builder.Spec] to return the final accumulated spec.
type Builder struct {
	spec      *aff.Spec                              // Accumulated state across all subsystems.
	provision *provision.Spec                        // Accumulated provision state.
	network   *net.Spec                              // Accumulated network spec state.
	volume    *volume.Spec                           // Accumulated volume declarations.
	kernel    *kernel.Spec                           // Accumulated kernel requirements.
	index     map[subsystem.Name]subsystem.Subsystem // Name-indexed dispatch map.
}

// Returns a [Builder].
//
// The OCI section of the spec starts with a deny-all baseline; subsystems can
// only loosen it. Non-OCI sections start in their zero-grant state.
func NewBuilder() *Builder {
	s := aff.NewSpec()
	prov := &provision.Spec{}
	netw := s.Net
	volm := s.Volume
	kspec := s.Kernel

	subs := []subsystem.Subsystem{
		cap.New(s.OCI.Process.Capabilities),
		rlimit.New(&s.OCI.Process.Rlimits),
		seccomp.New(s.OCI.Linux.Seccomp),
		fcap.New(s.Fcap),
		mac.New(s.MAC),
		cgroup.New(s.OCI.Linux.Resources),
		provision.New(prov),
		net.New(netw),
		mount.New(&s.OCI.Mounts),
		volume.New(volm),
		device.New(&s.OCI.Linux.Devices),
		kernel.New(kspec),
	}

	idx := make(map[subsystem.Name]subsystem.Subsystem, len(subs))
	for _, sub := range subs {
		idx[sub.Name()] = sub
	}

	return &Builder{spec: s, provision: prov, network: netw, volume: volm, kernel: kspec, index: idx}
}

// Returns the accumulated runtime spec produced by all processed grants.
func (b *Builder) Spec() *aff.Spec {
	return b.spec
}

// Returns the compute resource requirements accumulated from .provision grants.
func (b *Builder) Provision() *provision.Spec {
	return b.provision
}

// Returns the container network spec accumulated from .net grants.
func (b *Builder) Network() *net.Spec {
	return b.network
}

// Returns the persistent storage volumes accumulated from .volume grants.
func (b *Builder) Volumes() *volume.Spec {
	return b.volume
}

// Returns the kernel requirements accumulated from .kernel grants.
func (b *Builder) Kernel() *kernel.Spec {
	return b.kernel
}

// Processes a single grant, updating the accumulated spec.
//
// Reference grants are pulled from the registry via source and their grants
// are processed recursively. When the reference grant carries [manifest.Grant.Args],
// they are substituted into the pulled grants before recursing. Domain grants
// are dispatched to the matching subsystem. Returns [ErrResolution] for pull
// failures, parse errors, or unknown subsystem names.
func (b *Builder) Build(ctx context.Context, g manifest.Grant, src hub.Source) error {
	if g.IsRef() {
		a, _, err := pull(ctx, src, g.RefTarget())
		if err != nil {
			return crex.Wrapf(ErrResolution, err, "pull %s", g.RefTarget())
		}
		for _, scope := range a.Scopes {
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
		return crex.Wrapf(ErrResolution, err, "parse %q", g.Source)
	}
	return b.dispatch(parsed)
}

// Dispatches a parsed AGL model to the matching subsystem.
//
// Subsystems own repeat handling. They may treat repeated grants as a no-op,
// merge them into the accumulated state, or reject them as a conflict.
func (b *Builder) dispatch(p *agl.Model) error {
	name := subsystem.Name(p.Subsystem)
	sub, ok := b.index[name]
	if !ok {
		return crex.Newf(ErrUnknownSubsystem, "unknown subsystem %q", p.Subsystem)
	}
	return sub.Build(p)
}

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
// Resolves target as an affordance reference and pulls it via [hub.Source.Pull].
func pull(ctx context.Context, src hub.Source, target string) (*manifest.Affordance, string, error) {
	ref, err := src.Parse(string(manifest.TypeAffordance), target)
	if err != nil {
		return nil, "", err
	}
	result, err := src.Pull(ctx, ref)
	if err != nil {
		return nil, "", err
	}
	a, err := manifest.ReadAsAt[*manifest.Affordance](result.Extracted)
	if err != nil {
		return nil, "", crex.Wrapf(ErrResolution, err, "%s", target)
	}
	return a, result.Digest, nil
}
