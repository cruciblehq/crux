package manifest

import (
	"strings"

	"github.com/cruciblehq/crux/crex"
)

// Prefix that marks a grant source as a domain grant.
//
// Sources without the prefix are reference grants whose target is the source
// string. The prefix is not semantically meaningful and is stripped before
// parsing; it's only used to distinguish domain grants from reference grants
// in the manifest. The grammar for domain grants is defined in the affordance
// builder, which is responsible for parsing them.
const grantDomainPrefix = "."

// A single grant.
//
// Grants can appear in one of two forms. Reference grants name an affordance
// to be inlined during composition by the affordance builder; their source is
// a bare affordance name. Domain grants begin with [grantDomainPrefix] and are
// parsed by the affordance builder according to the AGL grammar. The prefix is
// not semantically meaningful and is stripped before parsing; it's only used
// to distinguish domain grants from reference grants in the manifest. Args are
// only valid on reference grants; each key-value pair is passed to the
// parameter named by the key.
type Grant struct {

	// Canonical source form of the grant.
	//
	// For domain grants, has the form parsed by the AGL grammar. For reference
	// grants, is a bare affordance name. Empty is invalid.
	Source string `codec:"-"`

	// Named arguments for a reference grant.
	//
	// Each key must match a parameter declared in the referenced affordance's
	// [Schema.Params]. Only valid on reference grants.
	Args Args `codec:"-"`
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

// Validates the grant source.
//
// The source must be non-empty, and args can only be present on reference
// grants. Syntax validation of domain grants and semantic validation against a
// specific subsystem happen later, during the build stage.
func (g *Grant) Validate() error {
	if g.Source == "" {
		return crex.Wrapf(ErrInvalidGrant, "empty grant")
	}
	if g.IsRef() {
		if err := g.Args.Validate(); err != nil {
			return crex.Wrap(ErrInvalidGrant, err)
		}
	} else {
		if len(g.Args) > 0 {
			return crex.Wrap(ErrInvalidGrant, ErrDomainGrantWithArgs)
		}
	}
	return nil
}

// Encodes the grant to its canonical serialized form.
//
// A grant with no args encodes to its source string. A grant with [Grant.Args]
// set encodes to a single-key map of source to a string-keyed map of arg values.
func (g *Grant) Encode() (any, error) {
	if len(g.Args) > 0 {
		args := make(map[string]any, len(g.Args))
		for k, v := range g.Args {
			args[k] = v
		}
		return map[string]any{g.Source: args}, nil
	}
	return g.Source, nil
}

// Decodes a raw grant element into the receiver.
//
// Strings are stored verbatim as [Grant.Source]. Maps must contain exactly one
// key. For domain grants, the value must be nil. For reference grants, a nil
// value means no args; a string-keyed map sets [Grant.Args].
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

// Populates the source and optional args from a single-entry map.
//
// Domain grants (key starts with ".") must have a nil value. Reference grants
// accept a nil value (no args) or a string-keyed map ([Grant.Args]).
func (g *Grant) decodeMap(m map[string]any) error {
	source, val, err := onlyKeyInGrantMap(m)
	if err != nil {
		return err
	}

	g.Source = source
	if !g.IsRef() {
		if val != nil {
			return crex.Wrapf(ErrInvalidGrant, "domain grant %q does not accept args", source)
		}
		return nil
	}
	return g.decodeRefArgs(source, val)
}

// Asserts that the given map has exactly one key and returns that key and
// its value.
func onlyKeyInGrantMap(m map[string]any) (string, any, error) {
	if len(m) != 1 {
		return "", nil, crex.Wrapf(ErrInvalidGrant, "grant must name exactly one source")
	}

	var source string
	var val any
	for source, val = range m {
		break
	}
	return source, val, nil
}

// Populates [Grant.Args] from the value of a reference grant.
//
// A nil value means no args. A string-keyed map sets [Grant.Args]; every
// value in the map must be a string.
func (g *Grant) decodeRefArgs(source string, val any) error {
	switch v := val.(type) {
	case nil:
		// no args
	case map[string]any:
		args := make(map[string]string, len(v))
		for k, av := range v {
			s, ok := av.(string)
			if !ok {
				return crex.Wrapf(ErrInvalidGrant, "arg %q of ref grant %q must be a string", k, source)
			}
			args[k] = s
		}
		g.Args = args
	default:
		return crex.Wrapf(ErrInvalidGrant, "unsupported arg type %T for grant %q", val, source)
	}
	return nil
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
