package subsystem

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
)

// Implements [Subsystem] for [manifest.DomainFcap].
//
// Accumulates file-capability grants into an internal model keyed by binary
// path. Grants for the same path are merged: capability lists are unified
// via [appendUnique] and the effective bit is OR'd. No conflicts are
// possible. File capabilities are purely additive.
type FcapsSubsystem struct {
	fcaps map[string]*fcap
}

// Feeds an fcap grant into the subsystem.
//
// Validates the expression "<verb> <path> <caps...>" and merges the result
// into the internal model. Returns the grant if it had an effect, or nil
// if the path already held all specified capabilities.
func (s *FcapsSubsystem) Build(_ context.Context, domain Domain, input manifest.Grant) ([]manifest.Grant, error) {
	crex.Assertf(domain == DomainFcap, "unexpected fcap domain %q", domain)

	fc, err := parseFcap(input.Expr)
	if err != nil {
		return nil, err
	}

	if !s.merge(&fc) {
		return nil, nil
	}

	return []manifest.Grant{{Subsystem: string(domain), Expr: input.Expr}}, nil
}

// Merges a parsed [fcap] into the accumulated model.
//
// If the path is new, the entry is stored directly. If the path already
// exists, capability lists are unified and the effective bit is OR'd.
// Returns true if the model changed.
func (s *FcapsSubsystem) merge(fc *fcap) bool {
	if s.fcaps == nil {
		s.fcaps = make(map[string]*fcap)
	}

	existing, ok := s.fcaps[fc.Path]
	if !ok {
		s.fcaps[fc.Path] = fc
		return true
	}

	changed := mergeSlice(&existing.Permitted, fc.Permitted)
	changed = mergeSlice(&existing.Inheritable, fc.Inheritable) || changed
	if !existing.Effective && fc.Effective {
		existing.Effective = true
		changed = true
	}
	return changed
}
