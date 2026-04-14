package subsystem

import (
	"context"
	"strings"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
)

// Implements [Subsystem] for [manifest.DomainCgroup].
//
// Accumulates cgroup grants into an internal [cgroup] model. Scalar knobs
// compare against their restrictive defaults: grants that equal the default
// are dropped (not a relaxation), identical relaxations are idempotent, and
// conflicting relaxations error. List knobs (e.g. io.max, hugetlb, device)
// always append.
type CgroupSubsystem struct {
	model cgroup
}

// Feeds a cgroup grant into the subsystem.
//
// Validates the expression "<knob> [value]" with optional sub-args and merges
// it into the internal model. Returns an error wrapping [ErrGrantConflict] if
// two grants set the same knob to different values.
func (s *CgroupSubsystem) Build(_ context.Context, domain Domain, input manifest.Grant) ([]manifest.Grant, error) {
	crex.Assertf(domain == DomainCgroup, "unexpected cgroup domain %q", domain)

	fields := strings.Fields(input.Expr)
	if len(fields) == 0 {
		return nil, crex.Wrapf(ErrGrantExpression, "knob required")
	}

	name := cgroupKnobName(fields[0])
	knob, ok := cgroupKnobs[name]
	if !ok {
		return nil, crex.Wrapf(ErrGrantExpression, "unknown knob %q", name)
	}

	val := strings.TrimSpace(strings.TrimPrefix(input.Expr, string(name)))
	applied, err := knob.apply(&s.model, val, input.Args)
	if err != nil {
		return nil, err
	}
	if !applied {
		return nil, nil
	}

	return []manifest.Grant{{Subsystem: string(domain), Expr: input.Expr, Args: input.Args}}, nil
}
