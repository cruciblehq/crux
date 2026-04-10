package subsystem

import (
	"context"

	"github.com/cruciblehq/crux/internal/manifest"
)

// Handles grants for one or more domains.
//
// Each implementor translates grant expressions into validated, expanded
// grants at build time and knows how to apply them to a runtime. A single
// type may handle several domains and uses the domain parameter to
// distinguish them.
type Subsystem interface {

	// Resolves a source grant into one or more built grants.
	//
	// Called at build time to validate and expand compact shorthand syntax.
	// A single input may expand into multiple grants (e.g. bracket expansion
	// in seccomp). Each returned grant has the same Subsystem as the input,
	// with Expr and Args validated and normalized.
	Build(ctx context.Context, domain Domain, input manifest.Grant) ([]manifest.Grant, error)

	// Applies a built grant to a runtime.
	//
	// Called at apply time to mutate the runtime according to a previously
	// built grant.
	Apply(ctx context.Context, domain Domain, grant manifest.Grant) error
}
