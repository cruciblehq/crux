// Package mac implements the mandatory access control subsystem.
//
// MAC grants target Linux Security Module hooks. Each grant names an LSM
// hook and may constrain it with a where clause built as an expression
// tree of comparisons, set membership tests, pattern matching, range
// checks, and bitwise tests combined with boolean operators. The hook
// catalog validates hook and field names so that grants referencing
// unknown hooks or fields are rejected at build time, and field types are
// checked against the operators applied to them.
//
// A [Subsystem] wraps a [spec.MACSpec] and accumulates grants in place.
// Build compiles one parsed grant into a hook allow rule and folds it into
// the spec with union semantics: rules are deduplicated by hook plus
// structurally identical where-clause, and rules for the same hook with
// differing where-clauses are kept because the enforcer OR's them.
//
//	s := mac.New(spec)
//	if err := s.Build(g); err != nil {
//		return err
//	}
package mac
