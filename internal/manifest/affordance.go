package manifest

import (
	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/codec"
)

// Holds configuration specific to affordance resources.
//
// An affordance declares grants that confer subsystem settings or reference
// other affordances. Domain grants use a dot-prefixed expression to select
// the subsystem (e.g. ".seccomp openat"). References name another affordance
// resource and are resolved recursively by AffordanceBuilder.
type Affordance struct {

	// Parameter schema for this affordance.
	//
	// Lists the named arguments the affordance accepts. When [Schema.Default]
	// is set, scalar values passed by a caller are assigned to that parameter
	// instead of requiring an explicit key. Zero value means no parameters.
	Schema Schema `codec:"schema,omitempty"`

	// Grant scopes.
	//
	// Internally groups grants by platform. Universal grants (no platform)
	// live in a scope with an empty Platform. Custom Encode flattens
	// universal grants into the top-level list; platform-scoped grants
	// are written as platform group entries.
	Scopes []GrantScope `codec:"-"`
}

// Validates the affordance configuration.
//
// Schema and all grant scopes must be valid.
func (a *Affordance) Validate() error {
	if err := a.Schema.Validate(); err != nil {
		return crex.Wrap(ErrInvalidAffordance, err)
	}

	for i := range a.Scopes {
		if err := a.Scopes[i].Validate(); err != nil {
			return crex.Wrapf(ErrInvalidAffordance, "grant %d: %w", i+1, err)
		}
	}

	return nil
}

// Encodes the affordance to a format-independent value.
//
// Implements [codec.Encodable]. Universal grants (scopes with empty Platform)
// are flattened into the top-level grants list. Platform-scoped grants are
// written as platform group entries with their own inner grants list.
func (a *Affordance) Encode() (any, error) {
	m := make(map[string]any)

	sm, err := codec.ToMap(a.Schema)
	if err != nil {
		return nil, err
	}
	if len(sm) > 0 {
		m["schema"] = sm
	}

	list, err := encodeScopes(a.Scopes)
	if err != nil {
		return nil, err
	}
	if len(list) > 0 {
		m["grants"] = list
	}

	return m, nil
}

// Encodes grant scopes into a flat list suitable for serialization.
//
// Universal grants (empty Platform) are inlined directly. Platform-scoped
// grants are written as platform group entries.
func encodeScopes(scopes []GrantScope) ([]any, error) {
	var list []any
	for _, scope := range scopes {
		encoded, err := scope.Encode()
		if err != nil {
			return nil, err
		}
		if entries, ok := encoded.([]any); ok {
			list = append(list, entries...)
		} else {
			list = append(list, encoded)
		}
	}
	return list, nil
}

// Decodes a raw parsed map into the affordance.
//
// Implements [codec.Decodable]. Each grant element in the list is decoded by
// [decodeGrant], which handles both source format (compact syntax strings and
// dot-prefixed domain maps) and resolved format (maps with subsystem/expr/args
// keys). Platform groups are decoded recursively.
func (a *Affordance) Decode(raw any) error {
	src, ok := raw.(map[string]any)
	if !ok {
		return crex.Wrapf(ErrInvalidAffordance, "expected map, got %T", raw)
	}

	if err := codec.Field(src, a, "Schema"); err != nil {
		return crex.Wrap(ErrInvalidAffordance, err)
	}

	list, _ := src["grants"].([]any)
	if len(list) == 0 {
		return nil
	}

	scopes, err := decodeScopes(list)
	if err != nil {
		return crex.Wrap(ErrInvalidAffordance, err)
	}
	a.Scopes = scopes
	return nil
}

// Decodes a grant list into scopes, handling platform groups.
//
// Strings and maps without a "platform" key are accumulated into the universal
// scope (empty Platform). Maps with a "platform" key are decoded as platform
// groups via [GrantScope.Decode].
func decodeScopes(list []any) ([]GrantScope, error) {
	var universal []Grant
	var scopes []GrantScope

	for _, elem := range list {
		m, isMap := elem.(map[string]any)
		if isMap {
			if _, ok := m["platform"]; ok {
				var scope GrantScope
				if err := scope.Decode(m); err != nil {
					return nil, err
				}
				scopes = append(scopes, scope)
				continue
			}
		}

		g, err := decodeGrant(elem)
		if err != nil {
			return nil, err
		}
		universal = append(universal, g)
	}

	if len(universal) > 0 {
		scopes = append([]GrantScope{{Grants: universal}}, scopes...)
	}
	return scopes, nil
}
