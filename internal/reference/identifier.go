package reference

import (
	"fmt"
	"strings"
)

// Resource identifier.
//
// An identifier locates a resource without specifying a particular version.
// Use [ParseIdentifier] to construct valid identifiers.
type Identifier struct {
	typ       string // Resource type (e.g., "widget"). Lowercase alphabetic only.
	registry  string // Registry host (e.g., "hub.cruciblehq.xyz:8080"). Empty when not specified.
	namespace string // Resource namespace. Empty when not specified.
	name      string // Resource name.
}

// Parses an identifier string.
//
// The contextType is an opaque string representing the resource type. It is
// used to set the type when the identifier string does not include one, or to
// validate it when it does. The reference package does not validate whether
// the type is known, leaving that up to the caller.
//
// The expected string format is:
//
//	[<type>] [[[registry/]namespace/]name]
//
// The type is optional and must be lowercase alphabetic. When omitted, the
// context type is used. When present, it must match the context type exactly.
//
// The resource location is a single token with up to three slash-separated
// segments:
//   - name: just the resource name
//   - namespace/name: namespace and resource name
//   - registry/namespace/name: registry, namespace, and resource name
//
// When segments are omitted, the corresponding fields are left empty in the
// returned identifier. Callers are expected to apply defaults where needed.
func ParseIdentifier(s string, contextType string) (*Identifier, error) {
	p := &identifierParser{
		tokens: strings.Fields(s),
	}
	return p.parse(contextType)
}

// Like [ParseIdentifier], but panics on error.
func MustParseIdentifier(s string, contextType string) *Identifier {
	id, err := ParseIdentifier(s, contextType)
	if err != nil {
		panic(err)
	}
	return id
}

// Creates a new identifier with all fields set.
func NewIdentifier(typ string, registry, namespace, name string) *Identifier {
	return &Identifier{
		typ:       typ,
		registry:  registry,
		namespace: namespace,
		name:      name,
	}
}

// Returns a copy of this identifier with defaults applied for any empty fields.
//
// If the registry is empty and defaultRegistry is non-empty, the registry is
// set. If the namespace is empty and defaultNamespace is non-empty, the
// namespace is set. Fields that are already populated are never overwritten.
func (id *Identifier) WithDefaults(defaultRegistry, defaultNamespace string) *Identifier {
	clone := *id
	if clone.registry == "" && defaultRegistry != "" {
		clone.registry = defaultRegistry
	}
	if clone.namespace == "" && defaultNamespace != "" {
		clone.namespace = defaultNamespace
	}
	return &clone
}

// Resource type (e.g., "widget"). Lowercase alphabetic only.
func (id *Identifier) Type() string {
	return id.typ
}

// Registry host. Empty when not specified in the parsed string.
func (id *Identifier) Registry() string {
	return id.registry
}

// Namespace segment. Empty when not specified in the parsed string.
func (id *Identifier) Namespace() string {
	return id.namespace
}

// Resource name.
func (id *Identifier) Name() string {
	return id.name
}

// Returns the path component.
//
// If both namespace and name are set, returns namespace/name. Otherwise
// returns just the name.
func (id *Identifier) Path() string {
	if id.namespace == "" {
		return id.name
	}
	return id.namespace + "/" + id.name
}

// Returns a string representation of the identifier.
//
// The output includes only the fields that are set: type is always present,
// registry and namespace are included only when non-empty. An identifier
// parsed without defaults will omit the registry and namespace even though
// they may be required for resolution.
func (id *Identifier) String() string {
	if id.registry != "" {
		return fmt.Sprintf("%s %s/%s", id.Type(), id.registry, id.Path())
	}
	return fmt.Sprintf("%s %s", id.Type(), id.Path())
}
