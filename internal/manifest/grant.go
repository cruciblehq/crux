package manifest

import (
	"strings"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/subsystem"
)

// Domain selector prefix.
//
// This is used as a prefix to distinguish domain grants from references in the
// compact YAML syntax. A grant string or map key starting with this prefix is
// a domain grant; anything else is a reference.
const domainPrefix = "."

// An element in an affordance's grant list.
//
// A leaf grant carries a subsystem-specific configuration in [Grant.Config].
// The Config type depends on the [Grant.Subsystem] domain. A group grant
// scopes children to a platform via [Grant.Platform] and [Grant.Grants]. The
// two forms are mutually exclusive. Invalid combinations are rejected.
type Grant struct {

	// Selects the subsystem domain targeted by the grant.
	//
	// Determines how [Grant.Config] is interpreted. For domain grants, this
	// is extracted from the dot-prefixed syntax (e.g. ".seccomp openat" for
	// DomainSeccomp). For references, this is [DomainRef] and Config holds
	// the reference target string. Ref grants are transient: they are
	// resolved during build and never appear in the build output. Must be
	// empty on group nodes.
	Subsystem Domain `codec:"subsystem,omitempty"`

	// Subsystem-specific configuration for this grant.
	//
	// The concrete type depends on [Grant.Subsystem]. For [DomainSeccomp],
	// this holds a [subsystem.SeccompRule]. For [DomainRef], this holds the
	// reference target string, which is resolved and flattened during build.
	Config any `codec:"config,omitempty"`

	// Restricts this grant or group to a specific platform.
	//
	// When set on a leaf grant, restricts it to the given platform. When set
	// with [Grant.Grants], creates a platform-scoped group; all children in
	// the group inherit the scope. The format is "os/arch" (e.g.
	// "linux/amd64"). Can be set on both leaf and group nodes.
	Platform string `codec:"platform,omitempty"`

	// Child grants scoped to the platform specified by [Grant.Platform].
	//
	// When set, leaf fields (Subsystem, Config) must be empty. Only one
	// level of nesting is permitted: children must be leafs and cannot
	// themselves contain children. Children follow the same rules as
	// top-level grants.
	Grants []Grant `codec:"grants,omitempty"`
}

// Validates the grant's structural integrity.
//
// Leaf fields (Subsystem, Config) and group field (Grants) are mutually
// exclusive. Platform can be set on either. Children of a group must be
// leaf grants, nested groups are not allowed.
func (g *Grant) Validate() error {
	hasLeaf := g.Subsystem != ""
	hasGrants := len(g.Grants) > 0

	if hasLeaf && hasGrants {
		return crex.Wrapf(ErrInvalidAffordance, "grant cannot have both subsystem domain and children")
	}

	if hasGrants {
		return g.validateChildren()
	}

	if !hasLeaf {
		return crex.Wrapf(ErrInvalidAffordance, "grant missing subsystem domain")
	}
	if g.Subsystem == DomainRef {
		if _, ok := g.Config.(string); !ok || g.Config.(string) == "" {
			return crex.Wrapf(ErrInvalidAffordance, "ref grant missing target")
		}
	}
	return nil
}

// Validates all children of a group grant.
//
// Children must be leaf grants, nested groups are not allowed.
func (g *Grant) validateChildren() error {
	for i := range g.Grants {
		if len(g.Grants[i].Grants) > 0 {
			return crex.Wrap(ErrInvalidAffordance, ErrNestedPlatformGroup)
		}
		if err := g.Grants[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Decodes a list of raw YAML elements into a []Grant.
func decodeGrantSlice(raw []any) ([]Grant, error) {
	var grants []Grant
	for _, elem := range raw {
		decoded, err := decodeGrant(elem)
		if err != nil {
			return nil, err
		}
		grants = append(grants, decoded...)
	}
	return grants, nil
}

// Classifies and decodes a single raw YAML element into one or more Grants.
//
// A string produces a single leaf grant. A map with a "platform" key is a
// platform group whose children are preserved as nested grants.
// Any other single-key map is a domain grant with structured args.
func decodeGrant(elem any) ([]Grant, error) {
	switch v := elem.(type) {
	case string:
		return decodeGrantString(v)
	case map[string]any:
		return decodeGrantMap(v)
	default:
		return nil, crex.Wrapf(ErrInvalidAffordance, "unsupported grant type %T", elem)
	}
}

// Decodes a string element into one or more Grants.
//
// Strings starting with [domainPrefix] are domain grants. All others
// are Crucible references.
func decodeGrantString(s string) ([]Grant, error) {
	if strings.HasPrefix(s, domainPrefix) {
		trimmed := s[len(domainPrefix):]
		domain, expr, _ := strings.Cut(trimmed, " ")
		return resolveGrants(Domain(domain), expr, nil)
	}
	return []Grant{{Subsystem: DomainRef, Config: s}}, nil
}

// Decodes a map element as either a platform group or a domain grant
// with structured args.
func decodeGrantMap(m map[string]any) ([]Grant, error) {
	if p, ok := m["platform"]; ok {
		ps, _ := p.(string)
		raw, _ := m["grants"].([]any)
		children, err := decodeGrantSlice(raw)
		if err != nil {
			return nil, err
		}
		return []Grant{{Platform: ps, Grants: children}}, nil
	}
	for key, val := range m {
		args, err := decodeArgs(val)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(key, domainPrefix) {
			trimmed := key[len(domainPrefix):]
			domain, expr, _ := strings.Cut(trimmed, " ")
			return resolveGrants(Domain(domain), expr, args)
		}
		return resolveGrants(DomainRef, key, args)
	}
	return nil, crex.Wrapf(ErrInvalidAffordance, "empty map grant")
}

// Resolves a compact grant expression into one or more leaf Grants.
//
// Dispatches to the appropriate subsystem resolver based on domain. A single
// expression may expand into multiple grants.
func resolveGrants(domain Domain, expr string, args []string) ([]Grant, error) {
	switch domain {
	case DomainSeccomp:
		rules, err := subsystem.ResolveSeccomp(expr, args)
		if err != nil {
			return nil, crex.Wrap(ErrInvalidAffordance, err)
		}
		grants := make([]Grant, len(rules))
		for i := range rules {
			grants[i] = Grant{Subsystem: domain, Config: rules[i]}
		}
		return grants, nil
	case DomainRef:
		return []Grant{{Subsystem: DomainRef, Config: expr}}, nil
	default:
		return nil, crex.Wrapf(ErrInvalidAffordance, "unsupported domain %q", domain)
	}
}

// Converts a map value to a string slice of args.
func decodeArgs(val any) ([]string, error) {
	list, ok := val.([]any)
	if !ok {
		return nil, nil
	}
	args := make([]string, 0, len(list))
	for _, a := range list {
		s, ok := a.(string)
		if !ok {
			return nil, crex.Wrapf(ErrInvalidAffordance, "arg must be a string, not %T", a)
		}
		args = append(args, s)
	}
	return args, nil
}
