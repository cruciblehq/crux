package manifest

import (
	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/codec"
)

// The canonical filename for Crucible resource manifests.
const ManifestFile = "crucible.yaml"

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
	return m.Config.(codec.Validatable).Validate()
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

	cfg, err := codec.ToMap(m.Config)
	if err != nil {
		return nil, crex.Wrap(ErrEncodeFailed, err)
	}

	for k, v := range cfg {
		base[k] = v
	}

	return base, nil
}

// Decodes a raw parsed map into the manifest.
//
// Implements [codec.Decodable]. The common fields are decoded first to
// determine [Resource.Type]. The raw map is then decoded into the concrete
// configuration type for that resource.
func (m *Manifest) Decode(raw map[string]any) error {
	if err := codec.Field(raw, m, "Version"); err != nil {
		return crex.Wrap(ErrDecodeFailed, err)
	}
	if err := codec.Field(raw, m, "Resource"); err != nil {
		return crex.Wrap(ErrDecodeFailed, err)
	}

	configs := map[ResourceType]any{
		TypeRuntime:    &Runtime{},
		TypeService:    &Service{},
		TypeWidget:     &Widget{},
		TypeTemplate:   &Template{},
		TypeAffordance: &Affordance{},
		TypeBlueprint:  &Blueprint{},
	}

	target, ok := configs[m.Resource.Type]
	if !ok {
		return crex.Wrap(ErrDecodeFailed, ErrInvalidResourceType)
	}

	if err := codec.Decode(raw, target); err != nil {
		return crex.Wrap(ErrDecodeFailed, err)
	}

	m.Config = target

	return nil
}
