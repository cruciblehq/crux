package runtime

import (
	"github.com/cruciblehq/crux/internal/crex"

	"github.com/cruciblehq/crux/internal/manifest/grant"
	"github.com/cruciblehq/crux/internal/runtime/cap"
	"github.com/cruciblehq/crux/internal/runtime/cgroup"
	"github.com/cruciblehq/crux/internal/runtime/fcap"
	"github.com/cruciblehq/crux/internal/runtime/mac"
	"github.com/cruciblehq/crux/internal/runtime/rlimit"
	"github.com/cruciblehq/crux/internal/runtime/seccomp"
	"github.com/cruciblehq/crux/internal/runtime/shared"
)

// Dispatches grant compilation across all registered subsystems.
//
// A Builder is created once per session. It owns the unified [shared.Spec]
// model and the implementations wired to their slices of the model. Callers
// feed parsed grants through Build and pre-compiled state through Merge.
// Build routes a single parsed grant to the correct subsystem based on the
// Subsystem field. Merge folds the matching section of another Spec into
// the builder's Spec by calling each subsystem's Merge method in a fixed
// order: subsystems that others may depend on run first. Spec returns the
// accumulated model once all input has been processed.
type Builder struct {
	spec  *shared.Spec                     // Accumulated state across all subsystems.
	subs  []shared.Subsystem               // All subsystem implementations in fixed order.
	index map[shared.Name]shared.Subsystem // Name-indexed dispatch map.
}

// Returns a new Builder with all subsystems wired to a new Spec.
//
// The unified Spec starts in the strictest possible state: the OCI section
// carries a deny-all baseline that subsystems can only loosen, and custom
// non-OCI sections start in their zero-grant state.
func NewBuilder() *Builder {
	s := shared.NewSpec()

	// Subsystems ordered by dispatch priority
	subs := []shared.Subsystem{
		cap.New(s.OCI.Process.Capabilities),
		rlimit.New(&s.OCI.Process.Rlimits),
		seccomp.New(s.OCI.Linux.Seccomp),
		fcap.New(s.Fcap),
		mac.New(s.MAC),
		cgroup.New(s.OCI.Linux.Resources),
	}

	// Indexed subsystems for dispatch by name
	idx := make(map[shared.Name]shared.Subsystem, len(subs))
	for _, sub := range subs {
		idx[sub.Name()] = sub
	}

	return &Builder{spec: s, subs: subs, index: idx}
}

// Builds a grant into the accumulated state.
//
// [Grant.Subsystem] selects the appropriate subsystem. Unknown names produce
// ErrUnknownSubsystem. Errors from the subsystem are returned unwrapped.
func (b *Builder) Build(g *grant.Grant) error {
	sub, ok := b.index[shared.Name(g.Subsystem)]
	if !ok {
		return crex.Wrapf(ErrUnknownSubsystem, "unknown subsystem %q", g.Subsystem)
	}
	return sub.Build(*g)
}

// Folds another [shared.Spec] into the accumulated state.
//
// Each subsystem reads its corresponding section of src and folds it into
// its own. A nil src is a no-op. Merge stops at the first error.
func (b *Builder) Merge(src *shared.Spec) error {
	if src == nil {
		return nil
	}
	for _, sub := range b.subs {
		if err := sub.Merge(*src); err != nil {
			return err
		}
	}
	return nil
}

// Returns the accumulated [shared.Spec].
//
// The returned value is the same pointer the builder mutates; callers must
// not modify it concurrently. The zero-grant state of every section is the
// strictest possible policy for that subsystem.
func (b *Builder) Spec() *shared.Spec {
	return b.spec
}
