package codec

import "github.com/cruciblehq/crex"

// Implemented by types that support structural validation.
//
// Types implementing this interface can be validated via the package-level
// [Validate] function after decoding. Validation is always a separate step
// from decoding. This allows users to decode data that may be structurally
// invalid, and defer validation until they have the full object graph.
type Validatable interface {

	// Validates the receiver's fields and invariants.
	//
	// Returns nil when the value is structurally valid. Implementations
	// should check required fields, mutual exclusion constraints, and
	// value ranges. For composite types, Validate should recurse into
	// nested [Validatable] children.
	Validate() error
}

// Validates v by calling its [Validatable.Validate] method.
//
// Returns nil when the value is structurally valid. Implementations should
// check required fields, mutual exclusion constraints, and value ranges.
// Returns a programming error if v does not implement [Validatable].
func Validate(v any) error {
	val, ok := v.(Validatable)
	if !ok {
		return crex.ProgrammingError("validate failed", "type does not implement Validatable").Err()
	}
	return val.Validate()
}
