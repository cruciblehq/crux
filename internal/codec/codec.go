package codec

import (
	"encoding/json"

	"github.com/cruciblehq/crex"
	"github.com/go-viper/mapstructure/v2"
	"gopkg.in/yaml.v3"
)

// Struct tag name used by mapstructure for field mapping.
//
// All types in the spec module use codec:"name" to declare their serialized
// field names. This enables support for both JSON and YAML without needing
// separate tags. The same tag is also used by the custom Encoder and Decoder
// interfaces to identify types that need custom encoding logic.
const Tag = "codec"

// Implemented by types that need custom encoding logic.
//
// When the package-level [Encode] function receives a value that implements
// this interface, it delegates to the Encode method instead of using the
// default logic.
type Encoder interface {

	// Encodes the receiver to bytes in the given format.
	Encode(Format) ([]byte, error)
}

// Implemented by types that need custom decoding logic.
//
// When the package-level [Decode] function receives a value that implements
// this interface, it delegates to the Decode method instead of using the
// default logic.
type Decoder interface {

	// Decodes data in the given format into the receiver.
	Decode([]byte, Format) error
}

// Implemented by types that support validation.
type Validatable interface {

	// Validates the receiver and returns an error if invalid.
	Validate() error
}

// Converts v to bytes in the given format.
//
// If v implements [Encoder], its Encode method is called. Otherwise v is
// converted to a map via [ToMap] and the map is encoded.
func Encode(v any, f Format) ([]byte, error) {
	if enc, ok := v.(Encoder); ok {
		return enc.Encode(f)
	}
	m, err := ToMap(v)
	if err != nil {
		return nil, err
	}
	return encodeMap(m, f)
}

// Populates v from data in the given format.
//
// If v implements [Decoder], its Decode method is called. Otherwise the
// data is decoded into a map and the map is applied to v via [FromMap].
func Decode(data []byte, v any, f Format) error {
	if dec, ok := v.(Decoder); ok {
		return dec.Decode(data, f)
	}
	m, err := decodeMap(data, f)
	if err != nil {
		return err
	}
	return FromMap(m, v)
}

// Validates v by calling its [Validatable.Validate] method.
//
// Returns a programming error if v does not implement [Validatable].
func Validate(v any) error {
	val, ok := v.(Validatable)
	if !ok {
		return crex.ProgrammingError("validate failed", "type does not implement Validatable").Err()
	}
	return val.Validate()
}

// Converts a struct to a map[string]any.
//
// Field names are determined by the codec struct tag. Embedded structs with
// codec:",squash" are flattened into the parent map.
func ToMap(v any) (map[string]any, error) {
	var m map[string]any
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: Tag,
		Result:  &m,
		Squash:  true,
	})
	if err != nil {
		return nil, err
	}
	if err := d.Decode(v); err != nil {
		return nil, err
	}
	return m, nil
}

// Populates a struct from a map.
//
// Field names are matched by the codec struct tag. Weakly-typed input is
// enabled so string-to-int and similar conversions are handled automatically.
func FromMap(m any, v any) error {
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          Tag,
		Result:           v,
		Squash:           true,
		WeaklyTypedInput: true,
		DecodeHook:       mapstructure.TextUnmarshallerHookFunc(),
	})
	if err != nil {
		return err
	}
	return d.Decode(m)
}

// Serializes a map[string]any to bytes using the standard library encoder for
// the given format.
func encodeMap(m map[string]any, f Format) ([]byte, error) {
	switch f {
	case JSON:
		return json.Marshal(m)
	case YAML:
		return yaml.Marshal(m)
	default:
		return nil, ErrUnsupportedFormat
	}
}

// Deserializes bytes into a map[string]any using the standard library decoder
// for the given format.
func decodeMap(data []byte, f Format) (map[string]any, error) {
	var m map[string]any
	switch f {
	case JSON:
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
	case YAML:
		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, err
		}
	default:
		return nil, ErrUnsupportedFormat
	}
	return m, nil
}
