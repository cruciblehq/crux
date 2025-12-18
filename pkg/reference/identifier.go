package reference

import (
	"fmt"
	"strings"
)

const (

	// Default protocol scheme.
	DefaultScheme = "https"

	// Default registry authority.
	DefaultRegistry = "registry.crucible.net"

	// Default namespace for resources in the default registry.
	DefaultNamespace = "official"
)

// Resource identifier.
//
// An identifier locates a resource without specifying a particular version.
// Use [ParseIdentifier] to construct valid identifiers.
type Identifier struct {
	typ       string
	scheme    string
	registry  string
	namespace string
	name      string
	path      string
}

// Options for parsing identifiers.
type IdentifierOptions struct {
	DefaultScheme    string // Protocol scheme when not specified. Uses [DefaultScheme] if empty.
	DefaultRegistry  string // Registry authority when not specified. Uses [DefaultRegistry] if empty.
	DefaultNamespace string // Namespace when not specified. Uses [DefaultNamespace] if empty.
}

// Returns default identifier options.
func DefaultIdentifierOptions() *IdentifierOptions {
	return &IdentifierOptions{
		DefaultScheme:    DefaultScheme,
		DefaultRegistry:  DefaultRegistry,
		DefaultNamespace: DefaultNamespace,
	}
}

// Returns the scheme, using the package default if not set.
func (o *IdentifierOptions) scheme() string {
	if o != nil && o.DefaultScheme != "" {
		return o.DefaultScheme
	}
	return DefaultScheme
}

// Returns the registry, using the package default if not set.
func (o *IdentifierOptions) registry() string {
	if o != nil && o.DefaultRegistry != "" {
		return o.DefaultRegistry
	}
	return DefaultRegistry
}

// Returns the namespace, using the package default if not set.
func (o *IdentifierOptions) namespace() string {
	if o != nil && o.DefaultNamespace != "" {
		return o.DefaultNamespace
	}
	return DefaultNamespace
}

// Parses an identifier string.
//
// The context type is required, and used to set the type when the identifier
// string does not include one, or to validate the type when it does. When
// the identifier string includes a type, it must match the context type.
//
// The expected string format is:
//
//	[<type>] [[scheme://]registry/]<path>
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
// Options can be nil, in which case package defaults are used.
func ParseIdentifier(s string, contextType string, options *IdentifierOptions) (*Identifier, error) {
	p := &identifierParser{
		tokens:  strings.Fields(s),
		options: options,
	}
	return p.parse(contextType)
}

// Like [ParseIdentifier], but panics on error.
func MustParseIdentifier(s string, contextType string, options *IdentifierOptions) *Identifier {
	id, err := ParseIdentifier(s, contextType, options)
	if err != nil {
		panic(err)
	}
	return id
}

// Resource type (e.g., "widget"). Lowercase alphabetic only.
func (id *Identifier) Type() string {
	return id.typ
}

// Protocol scheme (e.g., "https").
func (id *Identifier) Scheme() string {
	return id.scheme
}

// Registry authority (e.g., "registry.crucible.net").
func (id *Identifier) Registry() string {
	return id.registry
}

// Namespace segment of the path. Only used with the default registry.
func (id *Identifier) Namespace() string {
	return id.namespace
}

// Resource name. Only used with the default registry.
func (id *Identifier) Name() string {
	return id.name
}

// Returns the full path component.
//
// For default registry references, returns namespace/name. For non-default
// registries, returns the stored path.
func (id *Identifier) Path() string {
	if id.path != "" {
		return id.path
	}
	return id.namespace + "/" + id.name
}

// Returns the full URI, including scheme, registry, and path.
func (id *Identifier) URI() string {
	return fmt.Sprintf("%s://%s/%s", id.Scheme(), id.Registry(), id.Path())
}

// Returns the canonical string representation.
//
// The output always includes the type. The scheme and registry are always
// included, even when using defaults.
func (id *Identifier) String() string {
	return fmt.Sprintf("%s %s", id.Type(), id.URI())
}
