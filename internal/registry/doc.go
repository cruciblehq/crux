// Package registry defines types and interfaces for the Crucible artifact registry.
//
// The registry stores versioned artifacts organized into namespaces and
// resources. A namespace groups related resources, each of which can have
// multiple immutable versions and mutable channels. Channels act as named
// pointers to versions (e.g., "stable", "latest"), letting consumers track a
// release stream without pinning to a specific version number.
//
// Every entity has three type variants. Info types carry mutable fields and are
// used in create and update requests. Summary types include statistics and
// metadata and appear in list responses and nested contexts. Full types include
// the complete nested data and are used in single-entity detail responses.
//
// The [Registry] interface defines the full set of CRUD operations. [Client]
// implements it as an HTTP client against the Hub API, using vendor-specific
// media types (application/vnd.crucible.{name}.v0) in Content-Type and Accept
// headers. [SQLRegistry] implements it with a SQL database and file-based
// archive storage, using file locks for safe concurrent access.
//
// [Resolve] maps a [reference.Reference] to a concrete version by resolving
// either a channel to its pointed-to version or a semver constraint to the
// highest matching version.
package registry
