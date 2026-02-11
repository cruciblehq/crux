package reference

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/resource"
)

// Resource identifier.
//
// An identifier locates a resource without specifying a particular version.
// Use [ParseIdentifier] to construct valid identifiers.
type Identifier struct {
	typ       resource.Type
	registry  *url.URL
	namespace string
	name      string
	path      string
}

// Options for parsing identifiers.
type IdentifierOptions struct {
	DefaultRegistry  string // Registry authority when not specified.
	DefaultNamespace string // Namespace when not specified.
}

// Creates a new [IdentifierOptions] with the given defaults.
//
// Both parameters are required. Returns an error if either is empty.
func NewIdentifierOptions(defaultRegistry, defaultNamespace string) (IdentifierOptions, error) {
	if defaultRegistry == "" {
		return IdentifierOptions{}, ErrMissingDefaultRegistry
	}
	if defaultNamespace == "" {
		return IdentifierOptions{}, ErrMissingDefaultNamespace
	}
	return IdentifierOptions{
		DefaultRegistry:  defaultRegistry,
		DefaultNamespace: defaultNamespace,
	}, nil
}

// Parses an identifier string.
//
// The context type is required, and used to set the type when the identifier
// string does not include one, or to validate the type when it does. When
// the identifier string includes a type, it must match the context type.
//
// The expected string format is:
//
//	[<type>] [[scheme://]host/]<path>
//
// The type is optional and must be lowercase alphabetic. When omitted, the
// context type is used. When present, it must match the context type exactly.
//
// The resource location can take three forms:
//   - Full URI with scheme: https://registry.example.com/path/to/resource
//   - Registry without scheme: registry.example.com/path/to/resource
//   - Default registry path: namespace/name or just name
//
// When using the default registry, the namespace defaults to the configured
// default namespace.
func ParseIdentifier(s string, contextType resource.Type, options IdentifierOptions) (*Identifier, error) {
	p := &identifierParser{
		tokens:  strings.Fields(s),
		options: options,
	}
	id, err := p.parse(contextType)
	if err != nil {
		return nil, crex.Wrap(ErrInvalidIdentifier, err)
	}
	return id, nil
}

// Like [ParseIdentifier], but panics on error.
func MustParseIdentifier(s string, contextType resource.Type, options IdentifierOptions) *Identifier {
	id, err := ParseIdentifier(s, contextType, options)
	if err != nil {
		panic(err)
	}
	return id
}

// Creates a new identifier.
//
// The registry string is parsed as a URL. Returns an error if parsing fails.
func NewIdentifier(typ resource.Type, registry, namespace, name string) (*Identifier, error) {
	u, err := url.Parse(registry)
	if err != nil {
		return nil, crex.Wrap(ErrInvalidIdentifier, err)
	}
	return &Identifier{
		typ:       typ,
		registry:  u,
		namespace: namespace,
		name:      name,
		path:      "",
	}, nil
}

// Like [NewIdentifier], but panics on error.
func MustNewIdentifier(typ resource.Type, registry, namespace, name string) *Identifier {
	id, err := NewIdentifier(typ, registry, namespace, name)
	if err != nil {
		panic(err)
	}
	return id
}

// Resource type (e.g., "widget"). Lowercase alphabetic only.
func (id *Identifier) Type() resource.Type {
	return id.typ
}

// Registry URL.
//
// Returns a copy; callers cannot mutate the identifier.
func (id *Identifier) Registry() url.URL {
	return *id.registry
}

// Registry host authority, including port if present.
func (id *Identifier) Host() string {
	return id.registry.Host
}

// Registry hostname, without port.
func (id *Identifier) Hostname() string {
	return id.registry.Hostname()
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
	if id.namespace == "" {
		return id.name
	}
	return id.namespace + "/" + id.name
}

// Returns the full URI, including registry and path.
func (id *Identifier) URI() string {
	return fmt.Sprintf("%s/%s", id.registry.String(), id.Path())
}

// Returns the canonical string representation.
//
// The output always includes the type. The scheme and registry are always
// included, even when using defaults.
func (id *Identifier) String() string {
	return fmt.Sprintf("%s %s", id.Type(), id.URI())
}
