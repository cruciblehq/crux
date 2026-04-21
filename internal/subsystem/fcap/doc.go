// Package fcap implements the file capabilities subsystem.
//
// Fcap grants declare which Linux capabilities are set on specific binaries
// inside the container via the security.capability extended attribute. Each
// grant names a path and one or more capabilities to grant. The package
// provides parsing, validation, and compilation to file capability state.
package fcap
