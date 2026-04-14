package subsystem

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
)

// Implements [Subsystem] for [manifest.DomainCap].
//
// Accumulates capability grants into an internal [caps] model. Capabilities
// are purely additive. Granting the same capability twice is harmless.
type CapsSubsystem struct {
	model caps
}

// Feeds a caps grant into the subsystem.
//
// Validates the expression "[verb] <name>" and merges it into the internal
// model. Returns a single grant with the validated expression.
func (s *CapsSubsystem) Build(_ context.Context, domain Domain, input manifest.Grant) ([]manifest.Grant, error) {
	crex.Assertf(domain == DomainCap, "unexpected cap domain %q", domain)

	caps, err := parseCaps(input.Expr)
	if err != nil {
		return nil, err
	}

	s.model.merge(caps)

	return []manifest.Grant{{Subsystem: string(domain), Expr: input.Expr}}, nil
}
