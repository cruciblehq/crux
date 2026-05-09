package manifest

import (
	"github.com/cruciblehq/crux/crex"
)

// Identifies a Crucible resource by its reference string.
//
// A ref always carries a [Ref.Target] that names the resource. The optional
// [Ref.ID] field names a specific instance that persists through the
// deployment pipeline. Affordance refs may carry a scalar [Ref.Value] or
// an argument map [Ref.Args] to parameterize the referenced affordance.
type Ref struct {

	// Stable identifier for this instance.
	//
	// Set when the reference names a specific instance (e.g. a service in a
	// blueprint). The ID persists from composition through plan into state.
	ID string `json:"id,omitempty"`

	// Crucible reference string identifying the resource.
	Target string `json:"ref"`

	// Scalar parameter for the referenced affordance. Mutually exclusive
	// with [Ref.Args].
	Value string `json:"value,omitempty"`

	// Argument map for the referenced affordance. Mutually exclusive with
	// [Ref.Value].
	Args map[string]string `json:"args,omitempty"`
}

// Validates the ref.
func (r *Ref) Validate() error {
	if r.Target == "" {
		return crex.Wrap(ErrInvalidRef, ErrMissingRefTarget)
	}
	if r.Value != "" && len(r.Args) > 0 {
		return crex.Wrapf(ErrInvalidRef, "%w: %s", ErrRefMixed, r.Target)
	}
	return nil
}
