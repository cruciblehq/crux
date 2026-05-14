package manifest

import (
	"github.com/cruciblehq/crux/codec"
	"github.com/cruciblehq/crux/crex"
)

// Defines a Crucible resource.
//
// A manifest specifies metadata about the resource and its type-specific
// configuration. The [Manifest.Config] field is polymorphic, its type being
// determined by [Resource.Type]. Each resource has its own config type.
type Manifest struct {

	// Schema version of the manifest format.
	//
	// Determines how the rest of the manifest is interpreted. Currently
	// the only supported version is 0.
	Version int `codec:"version"`

	// Common metadata shared across all resource types.
	//
	// Includes the resource type, qualified name, and version. This is
	// required and must be valid for the manifest to be considered valid.
	Resource Resource `codec:"resource"`

	// Type-specific configuration.
	//
	// The concrete type depends on [Resource.Type]: [Runtime] from runtimes,
	// [Service] for services, [Widget] for widgets, etc.
	Config any `codec:"-"`
}

// Validates the manifest.
//
// The version must be 0. Resource metadata must be valid. Config must be
// present and match the resource type. The config is validated according
// to its concrete type.
func (m *Manifest) Validate() error {
	if m.Version != 0 {
		return crex.Wrap(ErrInvalidManifest, ErrUnsupportedVersion)
	}

	if err := m.Resource.Validate(); err != nil {
		return crex.Wrap(ErrInvalidManifest, err)
	}

	if m.Config == nil {
		return crex.Wrap(ErrInvalidManifest, ErrMissingConfig)
	}

	if err := m.validateConfig(); err != nil {
		return crex.Wrap(ErrInvalidManifest, err)
	}

	return nil
}

// Validates that Config matches the resource type and is internally valid.
//
// Checks that Config is the expected concrete type for the resource type. Then
// calls the config's Validate method to check its internal consistency.
func (m *Manifest) validateConfig() error {
	var match bool
	switch m.Resource.Type {
	case TypeRuntime:
		_, match = m.Config.(*Runtime)
	case TypeService:
		_, match = m.Config.(*Service)
	case TypeWidget:
		_, match = m.Config.(*Widget)
	case TypeTemplate:
		_, match = m.Config.(*Template)
	case TypeAffordance:
		_, match = m.Config.(*Affordance)
	case TypeBlueprint:
		_, match = m.Config.(*Blueprint)
	default:
		return ErrInvalidResourceType
	}
	if !match {
		return ErrConfigTypeMismatch
	}
	v, ok := m.Config.(codec.Validatable)
	if !ok {
		return crex.Wrapf(ErrInvalidManifest, "config type %T is not validatable", m.Config)
	}
	return v.Validate()
}

// Encodes the manifest to a serializable map.
//
// Implements [codec.Encodable]. [Manifest.Config] is merged into the base
// fields so that the output matches the flat canonical manifest format.
func (m *Manifest) Encode() (any, error) {
	base, err := codec.ToMap(m)
	if err != nil {
		return nil, crex.Wrap(ErrEncodeFailed, err)
	}

	cfg, err := encodeToMap(m.Config)
	if err != nil {
		return nil, crex.Wrap(ErrEncodeFailed, err)
	}

	return mergeMap(cfg, base)
}

// Decodes a raw parsed map into the manifest.
//
// Implements [codec.Decodable]. The common fields are decoded first to
// determine [Resource.Type]. The raw map is then decoded into the concrete
// configuration type for that resource.
func (m *Manifest) Decode(raw any) error {
	src, ok := raw.(map[string]any)
	if !ok {
		return crex.Wrapf(ErrDecodeFailed, "unexpected type %T", raw)
	}
	if err := codec.Field(src, m, "Version"); err != nil {
		return crex.Wrap(ErrDecodeFailed, err)
	}
	if err := codec.Field(src, m, "Resource"); err != nil {
		return crex.Wrap(ErrDecodeFailed, err)
	}

	var target any
	switch m.Resource.Type {
	case TypeRuntime:
		target = &Runtime{}
	case TypeService:
		target = &Service{}
	case TypeWidget:
		target = &Widget{}
	case TypeTemplate:
		target = &Template{}
	case TypeAffordance:
		target = &Affordance{}
	case TypeBlueprint:
		target = &Blueprint{}
	default:
		return crex.Wrap(ErrDecodeFailed, ErrInvalidResourceType)
	}

	if err := codec.Decode(src, target); err != nil {
		return crex.Wrap(ErrDecodeFailed, err)
	}

	m.Config = target

	return nil
}
