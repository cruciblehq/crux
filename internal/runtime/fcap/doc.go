// Package fcap implements the file capabilities subsystem.
//
// Fcap grants declare Linux capabilities that should be set on specific
// binaries inside the container via the security.capability extended
// attribute. Each grant names a capability, a mode that selects which set
// the capability lands in when the binary executes, and the absolute path
// of the target binary. Capability names are validated against the same
// catalog used by the cap subsystem so that the two agree on what the
// kernel exposes.
//
// A [Subsystem] wraps a [fcap.Spec] and accumulates grants in place. Build
// folds one parsed grant into the spec; Merge folds another spec's entries
// in. Both operations have union semantics: per-binary capability lists
// are deduplicated and the effective bit is OR'd, so a path that already
// has capabilities granted simply gains the new ones.
//
//	s := fcap.New(spec)
//	if err := s.Build(g); err != nil {
//		return err
//	}
package fcap
