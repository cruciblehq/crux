package subsystem

import (
	"context"

	"github.com/cruciblehq/crux/internal/manifest"
)

// Handles grants for one or more domains.
//
// Each implementor is a stateful accumulator: grants are fed one at a time
// via [Subsystem.Build], which validates, expands, and merges each grant
// into the subsystem's internal model. Conflicts with previously applied
// grants are detected at insertion time. A single type may handle several
// domains and uses the domain parameter to distinguish them.
type Subsystem interface {

	// Feeds a grant into the subsystem.
	//
	// Parses, validates, and expands compact shorthand syntax (e.g., bracket
	// expansion in seccomp). The expanded grants are merged into the internal
	// model. If the grant conflicts with a previously applied grant, an error
	// wrapping [ErrGrantConflict] is returned. The returned slice contains
	// the expanded, normalized grants for persistence as rules.
	Build(ctx context.Context, domain Domain, input manifest.Grant) ([]manifest.Grant, error)
}
