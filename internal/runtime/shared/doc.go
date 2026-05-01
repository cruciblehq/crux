// Package shared provides types and constants used across runtime subsystems.
//
// Types that belong to the Linux kernel's external interface rather than to
// any single subsystem live here, allowing subsystem packages to reference
// kernel concepts without importing one another. The capability catalog and
// its parser live in this package because both cap and fcap need them, and
// keeping the catalog here prevents one subsystem from owning a name space
// that the other depends on.
//
// Parsing a capability name from its lowercase form:
//
//	c, err := shared.ParseCap("net_admin")
package shared
