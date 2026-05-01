// Package macspec holds the data types for the MAC subsystem.
//
// Only data types and pure data manipulation helpers live here so the runtime
// spec can carry a [Spec] without pulling in the build-side translation logic.
// The hook catalog and the Build and Merge logic live in internal/runtime/mac.
package macspec
