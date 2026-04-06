// Package registry defines the data types, interface, and validation rules for
// the Crucible artifact registry.
//
// The registry organizes versioned artifacts into a three-level hierarchy:
// namespaces contain resources, resources contain versions. Channels provide
// named pointers to specific versions within a resource. Versions may be
// published (immutable) or unpublished (mutable), and each version can hold a
// compressed archive artifact.
//
// Each entity in the hierarchy is represented by four types. Info types carry
// mutable fields for create and update requests. Summary types provide
// lightweight metadata for list responses. Full types include the complete
// entity with nested summaries of children. List types wrap a slice of
// summaries with a count. For example, [NamespaceInfo] carries the fields for
// creating or updating a namespace, [NamespaceSummary] is the compact form
// returned inside lists, [Namespace] is the full representation with embedded
// resource summaries, and [NamespaceList] wraps a slice of summaries. The
// [Error] type carries machine-readable [ErrorCode] values alongside
// human-readable messages for API error responses.
//
// Every wire type has a corresponding [MediaType] constant following the
// pattern application/vnd.crucible.{name}.v0, used in HTTP Content-Type and
// Accept headers for format negotiation.
//
// The [Registry] interface defines the full set of CRUD operations across all
// entity types, including archive upload and download. Both the HTTP client in
// crux and the SQL store in hub implement this interface.
//
// All types implement a Validate method that checks field constraints: name
// format, version string format, timestamp ordering, resource type, archive
// field consistency, digest format, and count bounds. The [Encode] and [Decode]
// codec functions call Validate automatically before encoding and after
// decoding, ensuring only valid data crosses serialization boundaries.
//
// Creating a namespace through the registry interface:
//
//	ns, err := reg.CreateNamespace(ctx, registry.NamespaceInfo{
//	    Name:        "my-namespace",
//	    Description: "Example namespace",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Encoding a type to JSON with automatic validation:
//
//	data, err := registry.Encode(&registry.NamespaceInfo{
//	    Name:        "my-namespace",
//	    Description: "Example namespace",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Decoding JSON back into a typed value:
//
//	info, err := registry.Decode[registry.NamespaceInfo](data)
//	if err != nil {
//	    log.Fatal(err)
//	}
package registry
