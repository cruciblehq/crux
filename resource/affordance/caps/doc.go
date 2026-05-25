// Package caps provides the Linux capability catalog.
//
// Linux capabilities appear in more than one affordance subsystem: cap grants
// assign capabilities to processes, and fcap grants assign them to file
// extended attributes. This package defines the shared catalog so that both
// subsystems can validate names and normalise values.
//
// Parsing a capability name from its lowercase short form:
//
//	c, err := caps.Parse("net_admin")
package caps
