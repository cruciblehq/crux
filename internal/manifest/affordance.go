package manifest

import "github.com/cruciblehq/crex"

// Holds configuration specific to affordance resources.
//
// An affordance declares a set of grants that confer subsystem settings or
// reference other affordances. Domain grants use a dot-prefixed expression
// to select the subsystem (e.g. ".seccomp openat"). References name another
// affordance resource and are resolved recursively by AffordanceBuilder.
// Platform groups scope grants to a target platform and are preserved as
// nested [Grant] nodes with a single level of children.
type Affordance struct {

	// Parameter schema for this affordance.
	//
	// Lists the named arguments the affordance accepts. When [Schema.Default]
	// is set, scalar values passed by a caller are assigned to that parameter
	// instead of requiring an explicit key. Zero value means no parameters.
	Schema Schema `codec:"schema,omitempty"`

	// Sandbox grants that compose this affordance.
	//
	// Each grant targets a subsystem with an expression. Domain grants use
	// dot-prefixed syntax (".seccomp openat") and are decoded with the
	// subsystem extracted from the prefix. References use bare Crucible
	// references and are decoded with subsystem [SubRef]; they are resolved
	// recursively by AffordanceBuilder. Platform groups in the YAML are
	// preserved as nested [Grant] nodes, each group carrying a [Grant.Platform]
	// selector and a [Grant.Grants] list of children.
	Grants []Grant `codec:"grants,omitempty"`
}

// Validates the affordance configuration.
func (a *Affordance) Validate() error {
	if err := a.Schema.Validate(); err != nil {
		return crex.Wrap(ErrInvalidAffordance, err)
	}

	for i := range a.Grants {
		if err := a.Grants[i].Validate(); err != nil {
			return crex.Wrapf(ErrInvalidAffordance, "grant %d: %w", i+1, err)
		}
	}

	return nil
}
