package manifest

import (
	"strings"

	"github.com/cruciblehq/crex"
)

// An element in an affordance's grant list.
//
// Fields are either leaf or group fields. Leaf fields are [Grant.Domain],
// [Grant.Expr], [Grant.Args]; they describe a single grant. Group field
// [Grant.Grants] holds children. The two types are mutually exclusive.
// [Grant.Platform] can be set on either a leaf (scoping that single grant)
// or a group (scoping all children). Invalid combinations are rejected.
type Grant struct {

	// Selects the subsystem domain targeted by the grant.
	//
	// Domain grants use a dot-prefixed syntax (e.g. ".seccomp openat") and
	// are decoded with the subsystem corresponding to the domain extracted
	// from the prefix. References use bare names ("fd/dup") and are decoded
	// with [DomainRef]; they are resolved recursively by AffordanceBuilder
	// before reaching any runtime dispatcher. Must be empty on group nodes.
	Subsystem Domain `codec:"-"`

	// The expression payload passed to the subsystem handler.
	//
	// For domain grants, this is the text after the dot-prefixed subsystem
	// domain selector (e.g. "openat" from ".seccomp openat"). For references,
	// this is the Crucible reference. Must be empty on group nodes.
	Expr string `codec:"-"`

	// Structured arguments from the map form of a grant.
	//
	// When a YAML grant is written as a single-key map whose value is a list
	// of strings, each string becomes an element of Args. Each string is a
	// "key [value]" pair interpreted by the subsystem handler. Nil for bare
	// string grants. Must be nil on group nodes.
	Args []string `codec:"-"`

	// Restricts this grant or group to a specific platform.
	//
	// When set on a leaf grant, restricts it to the given platform. When set
	// with [Grant.Grants], creates a platform-scoped group; all children in
	// the group inherit the scope. The format is "os/arch" (e.g.
	// "linux/amd64"). Can be set on both leaf and group nodes.
	Platform string `codec:"platform,omitempty"`

	// Child grants scoped to the platform specified by [Grant.Platform].
	//
	// When set, leaf fields (Subsystem, Expr, Args) must be empty. Only one
	// level of nesting is permitted: children must be leaf grants and cannot
	// themselves contain children. Children follow the same rules as
	// top-level grants.
	Grants []Grant `codec:"grants,omitempty"`
}

// Validates the grant's structural integrity.
//
// Leaf fields (Subsystem, Expr, Args) and group field (Grants) are mutually
// exclusive. Platform can be set on either. Children of a group must be
// leaf grants — nested groups are not allowed.
func (g *Grant) Validate() error {
	hasLeaf := g.Subsystem != ""
	hasGrants := len(g.Grants) > 0

	if hasLeaf && hasGrants {
		return crex.Wrapf(ErrInvalidAffordance, "grant cannot have both subsystem domain and children")
	}

	if hasGrants {
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

	if !hasLeaf {
		return crex.Wrapf(ErrInvalidAffordance, "grant missing subsystem domain")
	}
	if g.Subsystem == DomainRef && g.Expr == "" {
		return crex.Wrapf(ErrInvalidAffordance, "ref grant missing target")
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
		return []Grant{decodeGrantString(v)}, nil
	case map[string]any:
		return decodeGrantMap(v)
	default:
		return nil, crex.Wrapf(ErrInvalidAffordance, "unsupported grant type %T", elem)
	}
}

// Decodes a string element into a Grant, selecting the subsystem from syntax.
func decodeGrantString(s string) Grant {
	if strings.HasPrefix(s, ".") {
		trimmed := s[1:]
		subsystem, expr, _ := strings.Cut(trimmed, " ")
		return Grant{Subsystem: Domain(subsystem), Expr: expr}
	}
	return Grant{Subsystem: DomainRef, Expr: s}
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
		if strings.HasPrefix(key, ".") {
			trimmed := key[1:]
			subsystem, expr, _ := strings.Cut(trimmed, " ")
			return []Grant{{Subsystem: Domain(subsystem), Expr: expr, Args: args}}, nil
		}
		return []Grant{{Subsystem: DomainRef, Expr: key, Args: args}}, nil
	}
	return nil, crex.Wrapf(ErrInvalidAffordance, "empty map grant")
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
