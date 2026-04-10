package manifest

import (
	"strings"

	"github.com/cruciblehq/crex"
)

// Prefix that marks a grant string or map key as a domain grant.
//
// In the compact syntax, domain grants start with a dot followed by the
// subsystem name (e.g. ".seccomp openat"). Strings and map keys without
// this prefix are treated as affordance references.
const domainPrefix = "."

// A single permission granted by an affordance.
//
// Each grant targets one subsystem domain and carries an expression with
// optional arguments. The compact syntax form (e.g. ".seccomp openat") is
// parsed into Subsystem + Expr by [decodeGrant]. Grants with an empty
// Subsystem are references to other affordances. Grants with a non-empty
// Subsystem are domain grants that the builder resolves into built grants
// with subsystem-specific expressions and arguments. The builder ensures
// that all grants it produces are valid before including them in the build
// output manifest.
type Grant struct {

	// Selects the subsystem domain targeted by the grant.
	//
	// Determines how Expr and Args are interpreted. Each subsystem defines
	// its own domain syntax and the builder routes grants to the subsystem
	// based on this field. Empty for references. Non-empty for domain grants.
	Subsystem string `codec:"subsystem,omitempty"`

	// The grant expression.
	//
	// Contains the expression for the grant. For domain grants, this is the
	// subsystem-specific expression (e.g. "openat" for seccomp, "effective
	// NET_RAW" for cap). For references, this is the affordance reference.
	Expr string `codec:"expr,omitempty"`

	// Optional arguments that qualify the expression.
	//
	// Subsystem-specific argument strings. For seccomp, these are argument
	// filters (e.g. "0 eq 1"). For cgroup, these are sub-knob values. Nil
	// when no arguments are needed.
	Args []string `codec:"args,omitempty"`
}

// Validates the grant's structural integrity.
//
// A grant must have a non-empty Subsystem and Expr. Any additional validation
// is the responsibility of the subsystem that produced the grant. The builder
// must ensure that all grants produced by a subsystem are valid.
func (g *Grant) Validate() error {
	if g.Subsystem == "" {
		return crex.Wrapf(ErrInvalidAffordance, "built grant must have a subsystem")
	}
	if g.Expr == "" {
		return crex.Wrapf(ErrInvalidAffordance, "grant must have an expr")
	}
	return nil
}

// Decodes a map into the grant.
//
// Implements [codec.Decodable]. A map with a "subsystem" key is a resolved
// grant. A map with a dot-prefixed key is a source domain grant whose value
// is an optional args list. A map with a bare key is a reference.
func (g *Grant) Decode(raw any) error {
	m, ok := raw.(map[string]any)
	if !ok {
		return crex.Wrapf(ErrInvalidAffordance, "expected map, got %T", raw)
	}
	if sub, _ := m["subsystem"].(string); sub != "" {
		g.Subsystem = sub
		g.Expr, _ = m["expr"].(string)
		if rawArgs, ok := m["args"].([]any); ok {
			g.Args = make([]string, len(rawArgs))
			for i, a := range rawArgs {
				g.Args[i], _ = a.(string)
			}
		}
		return nil
	}
	for key, val := range m {
		if !strings.HasPrefix(key, domainPrefix) {
			g.Expr = key
			return nil
		}
		trimmed := key[len(domainPrefix):]
		domain, expr, _ := strings.Cut(trimmed, " ")
		if domain == "" {
			return crex.Wrapf(ErrInvalidAffordance, "empty domain in key %q", key)
		}
		args, err := decodeArgs(val)
		if err != nil {
			return err
		}
		g.Subsystem = domain
		g.Expr = expr
		g.Args = args
		return nil
	}
	return crex.Wrapf(ErrInvalidAffordance, "empty map grant")
}

// Decodes a raw grant element into a typed Grant.
//
// Handles both source and resolved formats. Source strings use the compact
// syntax: ".domain expr" for domain grants, bare names for references.
// Maps are delegated to [Grant.Decode].
func decodeGrant(elem any) (Grant, error) {
	switch v := elem.(type) {
	case string:
		return decodeGrantString(v)
	case map[string]any:
		var g Grant
		if err := g.Decode(v); err != nil {
			return Grant{}, err
		}
		return g, nil
	default:
		return Grant{}, crex.Wrapf(ErrInvalidAffordance, "unsupported grant type %T", elem)
	}
}

// Parses a compact syntax string into a Grant.
//
// Dot-prefixed strings are domain grants (e.g. ".seccomp openat" becomes
// Grant{Subsystem: "seccomp", Expr: "openat"}). Bare names are references
// (e.g. "my-affordance" becomes Grant{Expr: "my-affordance"}).
func decodeGrantString(s string) (Grant, error) {
	if !strings.HasPrefix(s, domainPrefix) {
		return Grant{Expr: s}, nil
	}
	trimmed := s[len(domainPrefix):]
	domain, expr, _ := strings.Cut(trimmed, " ")
	if domain == "" {
		return Grant{}, crex.Wrapf(ErrInvalidAffordance, "empty domain in %q", s)
	}
	return Grant{Subsystem: domain, Expr: expr}, nil
}

// Decodes args from a raw YAML value.
//
// Nil means no args. A []any of strings is converted to []string.
func decodeArgs(val any) ([]string, error) {
	if val == nil {
		return nil, nil
	}
	list, ok := val.([]any)
	if !ok {
		return nil, crex.Wrapf(ErrInvalidAffordance, "args must be a list, got %T", val)
	}
	args := make([]string, len(list))
	for i, a := range list {
		s, ok := a.(string)
		if !ok {
			return nil, crex.Wrapf(ErrInvalidAffordance, "arg %d must be a string, got %T", i+1, a)
		}
		args[i] = s
	}
	return args, nil
}
