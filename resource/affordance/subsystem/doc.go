// Package subsystem defines the dispatch contract shared by every affordance
// subsystem implementation.
//
// The Name type and the Subsystem interface live here so subsystem packages can
// reference the contract without importing one another. The orchestrator
// (Builder) uses these types to route parsed grants to the correct subsystem.
package subsystem
