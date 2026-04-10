package manifest

import (
	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/codec"
)

// Groups [Grant] values under a shared platform constraint.
//
// Platform selectors allow a resource to grants for different platforms. When
// a scope's Platform is empty, its grants apply universally. When non-empty,
// the format is "os/arch" (e.g. "linux/amd64"). Grant scopes are produced by
// the builder when resolving grants for an affordance and are written to the
// output manifest. They are also decoded when reading manifests (e.g. pulled
// affordances) and validated at apply time.
type GrantScope struct {

	// Platform selector for the grants in this scope.
	//
	// When empty, the grants apply to all platforms. When non-empty, the
	// format is "os/arch" (e.g. "linux/amd64") and the grants apply only
	// to matching platforms. The builder groups grants under scopes based
	// on their platform selectors.
	Platform string `codec:"platform,omitempty"`

	// Grants within this scope.
	//
	// Each grant targets a subsystem domain and carries its expression and
	// optional arguments. The builder produces scopes by grouping grants
	// with the same platform selector together.
	Grants []Grant `codec:"grants,omitempty"`
}

// Validates the scope.
//
// Every contained grant must itself be valid.
func (gs *GrantScope) Validate() error {
	for i := range gs.Grants {
		if err := gs.Grants[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Encodes the scope into one or more list entries.
//
// Implements [codec.Encodable]. Platform-scoped grants produce a single map
// with platform and grants keys. Universal grants (empty Platform) produce
// one map per grant, suitable for flattening into the parent list.
func (gs *GrantScope) Encode() (any, error) {
	if gs.Platform != "" {
		return codec.ToMap(*gs)
	}
	entries := make([]any, 0, len(gs.Grants))
	for _, g := range gs.Grants {
		gm, err := codec.ToMap(g)
		if err != nil {
			return nil, err
		}
		entries = append(entries, gm)
	}
	return entries, nil
}

// Decodes a platform group map into the scope.
//
// Implements [codec.Decodable]. The map must contain a "grants" key with a
// list of grant elements. Inner grants are decoded via [decodeGrant].
// Platform groups cannot be nested.
func (gs *GrantScope) Decode(raw any) error {
	src, ok := raw.(map[string]any)
	if !ok {
		return crex.Wrapf(ErrInvalidAffordance, "expected map, got %T", raw)
	}
	inner, ok := src["grants"].([]any)
	if !ok {
		return crex.Wrapf(ErrInvalidAffordance, "platform group missing grants key")
	}
	gs.Platform, _ = src["platform"].(string)
	for _, elem := range inner {
		g, err := decodeGrant(elem)
		if err != nil {
			return err
		}
		gs.Grants = append(gs.Grants, g)
	}
	return nil
}
