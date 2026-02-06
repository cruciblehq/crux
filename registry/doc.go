// Package registry defines types and interfaces for the Crucible artifact registry.
//
// The registry provides hierarchical storage for versioned artifacts organized
// into namespaces and resources. Each namespace contains multiple resources,
// and each resource can have multiple versions and channels. Versions are
// immutable once published, while channels provide mutable pointers to versions
// for dynamic references.
//
// The registry uses a three-level hierarchy:
//
//   - Namespace: Top-level organizational unit for grouping related resources
//   - Resource: Publishable artifact with multiple versions and channels
//   - Version: Immutable snapshot of a resource at a specific point in time
//   - Channel: Mutable pointer to a version (e.g., "stable", "latest")
//
// The package defines three variants of each entity type:
//
//   - Info types: Used in create and update requests, containing only mutable fields
//   - Summary types: Used in lists and nested contexts, including statistics and metadata
//   - Full types: Used in single-entity responses, including complete nested data
//
// All request and response bodies use vendor-specific media types following the
// pattern application/vnd.crucible.{name}.v0. Clients specify media types in
// Content-Type headers for requests and Accept headers for responses.
//
// Operations return errors with platform-specific error codes providing granular
// classification beyond HTTP status codes. Error responses use the Error type
// with machine-readable codes and human-readable messages.
package registry
