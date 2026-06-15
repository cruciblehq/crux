// Package capset provides the Linux capability catalog.
//
// Linux capabilities appear in two affordance subsystems: cap grants assign
// capabilities to processes and fcap grants assign them to file extended
// attributes. This package defines the shared catalog so that both subsystems
// can validate names and normalise values.
//
// Parsing a capability name from its lowercase short form:
//
//	c, err := capset.Parse("net_admin")
package capset
