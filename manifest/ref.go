package manifest

import (
	"github.com/cruciblehq/crux/crex"
)

// Identifies a resource by its reference string.
//
// A reference carries a [Ref.Target] used to look up and fetch the resource
// from the registry. [Ref.ID] assigns a stable identifier to the reference
// when it is used to name a specific instance of a resource (e.g. a service
// in a blueprint). The ID persists from composition through plan into state,
// allowing the instance to be tracked and referenced across the deployment
// lifecycle. The reference may also carry a scalar [Ref.Value] or an argument
// map [Ref.Args] to parameterize the referenced affordance. The two fields are
// mutually exclusive; the presence of both is an error.
type Ref struct {

	// Stable identifier for this instance.
	//
	// Set when the reference names a specific instance (e.g. a service in a
	// blueprint). The ID persists from composition through plan into state.
	ID string `codec:"id,omitempty"`

	// Crucible reference string identifying the resource.
	//
	// The format is "namespace/name version", where namespace/name is the
	// qualified resource name and version is a semver string. The registry
	// resolves this string to a specific artifact. This field is required.
	Target string `codec:"target"`

	// Scalar argument for the referenced affordance.
	//
	// Passes a value to the parameter named by [Schema.Default], without
	// requiring an explicit key. Other parameters declared by the affordance
	// may still exist as long as they are optional. Mutually exclusive with
	// [Ref.Args]; specifying both is an error.
	Value string `codec:"value,omitempty"`

	// Named arguments for the referenced affordance.
	//
	// Maps parameter names to their string values. Each key must match a
	// parameter declared in the affordance's [Schema.Params]. Mutually
	// exclusive with [Ref.Value]; specifying both is an error.
	Args Args `codec:"args,omitempty"`
}

// Validates the ref.
func (r *Ref) Validate() error {
	if r.ID != "" && !isValidName(r.ID) {
		return crex.Wrap(ErrInvalidRef, ErrInvalidRefID)
	}
	if r.Target == "" {
		return crex.Wrap(ErrInvalidRef, ErrMissingRefTarget)
	}
	if r.Value != "" && len(r.Args) > 0 {
		return crex.Wrapf(ErrInvalidRef, "%w: %s", ErrRefMixed, r.Target)
	}
	if err := r.Args.Validate(); err != nil {
		return crex.Wrap(ErrInvalidRef, err)
	}
	return nil
}
