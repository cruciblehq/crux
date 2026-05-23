// Package fio provides low-level file I/O utilities shared across the codebase.
//
// It covers two concerns. File locking: exclusive advisory locks via flock(2),
// with context-aware acquisition that honours cancellation and deadlines.
// File helpers: atomic writes with SHA-256 digest tracking, existence checks,
// empty-directory pruning, and subdirectory enumeration.
//
// All helpers operate on plain [os.File] values and filesystem paths; they
// carry no domain knowledge about manifests, caches, or any other Crucible
// concept.
package fio
