package manifest

import "github.com/cruciblehq/crux/reference"

// Source type for a runtime base image.
type RuntimeSourceType string

const (

	// Local OCI image archive.
	RuntimeSourceFile RuntimeSourceType = "file"

	// Crucible runtime reference.
	RuntimeSourceRef RuntimeSourceType = "ref"
)

// Describes the origin of a runtime's base image.
//
// Type discriminates between a local OCI tarball and a Crucible runtime
// reference. Value holds the raw payload after the type prefix. For ref
// sources, Ref is populated with the parsed [reference.Reference].
type RuntimeSource struct {

	// Discriminates between file and ref sources.
	Type RuntimeSourceType

	// Source payload after the type prefix.
	//
	// For file sources this is the local OCI image archive path relative to
	// the manifest file. For ref sources this is the Crucible runtime
	// reference string as passed to [reference.Parse].
	Value string

	// Parsed Crucible runtime reference.
	//
	// Set when Type is [RuntimeSourceRef]. The reference type is always
	// [resource.TypeRuntime].
	Ref *reference.Reference
}
