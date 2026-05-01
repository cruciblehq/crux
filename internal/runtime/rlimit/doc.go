// Package rlimit implements the POSIX resource limits subsystem.
//
// Rlimit grants declare per-process resource ceilings that the kernel
// enforces via setrlimit(2). Each grant names a resource and supplies a
// soft limit and an optional hard limit, with the hard limit defaulting
// to the soft limit when absent. Resource names are validated against the
// catalog of known POSIX rlimit identifiers, and conflicting grants for
// the same resource are reported as errors so that policy authors notice
// contradictions instead of silently keeping the last value seen.
//
// A [Subsystem] wraps the OCI rlimits slice header of the unified spec.
// The shared baseline pre-populates that slice with one entry per known
// resource at soft=hard=0; Build updates the matching entry in place
// rather than appending a duplicate, so resources never granted remain
// at zero quota. Merge folds another spec's rlimit entries in with the
// same conflict semantics.
//
//	s := rlimit.New(&spec.Process.Rlimits)
//	if err := s.Build(g); err != nil {
//		return err
//	}
package rlimit
