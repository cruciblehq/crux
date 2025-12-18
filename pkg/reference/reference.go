package reference

import (
	"strings"
)

// Resource reference.
//
// A reference encapsulates all information needed to locate, identify, and
// verify a Crucible resource. It combines an [Identifier] with version
// information. References are immutable once created. Use [Parse] to
// construct valid references.
type Reference struct {
	Identifier
	version *VersionConstraint
	channel *string
	digest  *Digest
}

// Parses a reference string.
//
// The context type is required, and used to set the type when the reference
// string does not include one, or to validate the type when it does. When
// the reference string includes a type, it must match the context type.
//
// The expected string format is:
//
//	[<type>] [[scheme://]registry/]<path> (<version> | <channel>) [<digest>]
//
// The type is optional and must be lowercase alphabetic. When omitted, the
// context type is used. When present, it must match the context type exactly.
//
// The resource location can take three forms:
//   - Full URI with scheme: https://registry.example.com/path/to/resource
//   - Registry without scheme: registry.example.com/path/to/resource
//   - Default registry path: namespace/name or just name
//
// When using the default registry, the namespace defaults to "official" if
// not specified. Registry detection relies on the presence of dots in the
// first path segment.
//
// Either a version constraint or a channel is required, but not both. Version
// constraints may span multiple tokens (e.g., ">=1.0.0 <2.0.0"). Channels are
// prefixed with a colon (e.g., ":stable").
//
// The digest is optional and follows the format algorithm:hash (e.g.,
// "sha256:abcd1234"). When present, it freezes the reference to a specific
// content version.
//
// Options can be nil, in which case package defaults are used.
func Parse(s string, contextType string, options *IdentifierOptions) (*Reference, error) {
	p := &referenceParser{
		tokens:  strings.Fields(s),
		options: options,
	}
	return p.parse(contextType)
}

// Like [Parse], but panics on error.
func MustParse(s string, contextType string, options *IdentifierOptions) *Reference {
	ref, err := Parse(s, contextType, options)
	if err != nil {
		panic(err)
	}
	return ref
}

// Semantic version constraint. Nil if channel-based.
func (r *Reference) Version() *VersionConstraint {
	return r.version
}

// Named release track. Nil if version-based.
func (r *Reference) Channel() *string {
	return r.channel
}

// Cryptographic hash for content verification. Nil if not frozen.
func (r *Reference) Digest() *Digest {
	return r.digest
}

// Whether the reference includes a digest.
//
// A frozen reference refers to an exact, immutable resource version.
func (r *Reference) IsFrozen() bool {
	return r.digest != nil
}

// Whether the reference uses a channel instead of a version constraint.
func (r *Reference) IsChannelBased() bool {
	return r.channel != nil
}

// Whether the reference uses a version constraint.
func (r *Reference) IsVersionBased() bool {
	return r.version != nil
}

// Returns the canonical string representation.
//
// The output always includes the type. The scheme and registry are always
// included, even when using defaults. The path is always included. For default
// registry references, the path corresponds to namespace/name. Version or
// channel is always included, and digest is appended if present.
func (r *Reference) String() string {
	var sb strings.Builder

	sb.WriteString(r.Identifier.String())

	if r.IsChannelBased() {
		sb.WriteString(" :")
		sb.WriteString(*r.channel)
	} else if r.IsVersionBased() {
		sb.WriteByte(' ')
		sb.WriteString(r.version.String())
	}

	if r.IsFrozen() {
		sb.WriteByte(' ')
		sb.WriteString(r.digest.String())
	}

	return sb.String()
}
