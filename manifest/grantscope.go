package manifest

import (
	"github.com/cruciblehq/crux/crex"
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
	// Each grant targets a subsystem and carries its expression. The builder
	// produces scopes by grouping grants with the same platform selector.
	Grants []Grant `codec:"grants,omitempty"`
}

// Validates the grant scope.
//
// Platform, when set, must use the os/arch format (e.g. "linux/amd64").
// Every contained grant must itself be valid.
func (gs *GrantScope) Validate() error {
	if gs.Platform != "" && !isValidPlatform(gs.Platform) {
		return crex.Wrap(ErrInvalidAffordance, ErrInvalidPlatform)
	}
	for i := range gs.Grants {
		if err := gs.Grants[i].Validate(); err != nil {
			return crex.At(err, "grant", i+1)
		}
	}
	return nil
}

// Encodes the grant scope into one or more list entries.
//
// Platform-scoped grants produce a single map with platform and grants keys.
// Universal grants (empty Platform) produce one map per grant, suitable for
// flattening into the parent list.
func (gs *GrantScope) Encode() (any, error) {
	if gs.Platform != "" {
		entries := make([]any, 0, len(gs.Grants))
		for i := range gs.Grants {
			ge, err := gs.Grants[i].Encode()
			if err != nil {
				return nil, err
			}
			entries = append(entries, ge)
		}
		return map[string]any{
			"platform": gs.Platform,
			"grants":   entries,
		}, nil
	}
	entries := make([]any, 0, len(gs.Grants))
	for i := range gs.Grants {
		ge, err := gs.Grants[i].Encode()
		if err != nil {
			return nil, err
		}
		entries = append(entries, ge)
	}
	return entries, nil
}

// Decodes a platform group map into the scope.
//
// The map must contain a "grants" key with a list of grant elements. Inner
// grants are decoded via [decodeGrant]. Platform groups cannot be nested.
func (gs *GrantScope) Decode(raw any) error {
	src, ok := raw.(map[string]any)
	if !ok {
		return crex.Newf(ErrInvalidAffordance, "unexpected type %T", raw)
	}
	inner, ok := src["grants"].([]any)
	if !ok {
		return crex.Newf(ErrInvalidAffordance, "platform group missing grants key")
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
