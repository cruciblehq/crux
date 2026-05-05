package manifest

import (
	"strings"

	"github.com/cruciblehq/crux/internal/aegis"
	"github.com/cruciblehq/crux/internal/crex"
)

// Prefix that marks a grant source as a domain grant.
//
// Sources without the prefix are reference grants whose target is the bare
// source string.
const grantDomainPrefix = "."

// A single Aegis grant.
//
// Holds the canonical Aegis source string for the grant. Grants begin with
// [grantDomainPrefix] and follow the Aegis grammar; callers that need the
// parsed AST call [Grant.Parse]. Reference grants do not begin with the
// prefix and name another affordance to be inlined during build time;
// their target is the source string itself, returned by [Grant.RefTarget].
// Callers that need the parsed AST repeatedly should cache the result of
// [Grant.Parse] themselves; the model deliberately does not.
type Grant struct {

	// Canonical source form of the grant.
	//
	// For domain grants, has the form parsed by the Aegis grammar. For
	// reference grants, is a bare affordance name. Empty is invalid.
	Source string `json:"-"`
}

// Whether the grant is a reference to another affordance.
//
// A reference is any source that does not begin with [grantDomainPrefix].
// Reference grants are inlined during composition by the affordance builder
// and never reach the runtime subsystem dispatcher.
func (g *Grant) IsRef() bool {
	return !strings.HasPrefix(g.Source, grantDomainPrefix)
}

// Returns the affordance target of a reference grant, or "" for domain grants.
func (g *Grant) RefTarget() string {
	if g.IsRef() {
		return g.Source
	}
	return ""
}

// Parses a domain grant into its full AST representation.
//
// Returns an error wrapping [ErrInvalidGrant] when the receiver is a reference
// grant or when the source string fails to parse. Callers that need the AST
// repeatedly should cache the returned value; the manifest model does not.
func (g *Grant) Parse() (*aegis.Model, error) {
	if g.IsRef() {
		return nil, crex.Wrapf(ErrInvalidGrant, "cannot parse reference grant %q", g.Source)
	}
	p, err := aegis.Parse(g.Source)
	if err != nil {
		return nil, crex.Wrap(ErrInvalidGrant, err)
	}
	return p, nil
}

// Validates the grant source.
//
// Reference grants must have a non-empty source. Domain grants must parse
// cleanly via [aegis.Parse]. Semantic validation against a specific subsystem
// happens later in the runtime subsystem builder.
func (g *Grant) Validate() error {
	if g.Source == "" {
		return crex.Wrapf(ErrInvalidGrant, "empty grant")
	}
	if g.IsRef() {
		return nil
	}
	if _, err := aegis.Parse(g.Source); err != nil {
		return crex.Wrap(ErrInvalidGrant, err)
	}
	return nil
}

// Encodes the grant to its canonical source string.
//
// Implements [codec.Encodable]. The returned value is always a string suitable
// for inclusion in a YAML or JSON grant list.
func (g *Grant) Encode() (any, error) {
	return g.Source, nil
}

// Decodes a raw grant element into the receiver.
//
// Implements [codec.Decodable]. Strings are stored verbatim. Maps must contain
// exactly one key whose value is nil; the key is stored as the source string.
// The leading "." is interpreted on demand by [Grant.IsRef], not here.
func (g *Grant) Decode(raw any) error {
	switch v := raw.(type) {
	case string:
		g.Source = v
		return nil
	case map[string]any:
		return g.decodeMap(v)
	default:
		return crex.Wrapf(ErrInvalidGrant, "unsupported grant type %T", raw)
	}
}

// Populates the source from a single-entry map.
func (g *Grant) decodeMap(m map[string]any) error {
	if len(m) != 1 {
		return crex.Wrapf(ErrInvalidGrant, "grant map must have exactly one key")
	}
	for key, val := range m {
		if val != nil {
			return crex.Wrapf(ErrInvalidGrant, "grant key %q does not accept a value", key)
		}
		g.Source = key
		return nil
	}
	return crex.Wrapf(ErrInvalidGrant, "empty map grant")
}

// Decodes a raw grant element into a typed Grant.
//
// Helper to share the element-decoding logic. Errors are wrapped with
// [ErrInvalidAffordance] so that affordance-level callers can match on
// a sentinel regardless of which sub-decoder produced the original failure.
func decodeGrant(elem any) (Grant, error) {
	var g Grant
	if err := g.Decode(elem); err != nil {
		return Grant{}, crex.Wrap(ErrInvalidAffordance, err)
	}
	return g, nil
}
