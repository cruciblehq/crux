package codec

import (
	"encoding/json"

	"github.com/go-viper/mapstructure/v2"
	"gopkg.in/yaml.v3"
)

// Implemented by types that need custom encoding logic.
//
// When the package-level [Encode] function encounters a value that implements
// this interface, it delegates to Encode instead of using the default encode
// algorithm. The returned value must be a map, slice, or scalar that the format
// encoder (JSON/YAML) can serialize.
type Encodable interface {

	// Encodes the receiver to a format-independent value.
	//
	// The returned value is passed directly to the format serializer. Typical
	// return types are map[string]any for struct-like data, []any for lists,
	// or a primitive. Implementations may call [ToMap] on embedded structs
	// that don't need custom logic, then merge or transform the result before
	// returning. Returning an error aborts the outer [Encode] call.
	Encode() (any, error)
}

// Converts v to bytes in the given format.
//
// If v implements [Encodable], its Encode method provides the serializable
// representation. Otherwise v is converted via [ToMap]. The resulting value
// is then serialized to the requested format.
func Encode(v any, f Format) ([]byte, error) {
	var raw any
	var err error
	if enc, ok := v.(Encodable); ok {
		raw, err = enc.Encode()
	} else {
		raw, err = ToMap(v)
	}
	if err != nil {
		return nil, err
	}
	return encodeValue(raw, f)
}

// Converts a struct to a map[string]any.
//
// Field names are determined by the codec struct tag. Embedded structs with
// tag:",squash" are flattened into the parent map.
func ToMap(v any) (map[string]any, error) {
	var m map[string]any
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: tag,
		Result:  &m,
		Squash:  true,
		Deep:    true,
	})
	if err != nil {
		return nil, err
	}
	if err := d.Decode(v); err != nil {
		return nil, err
	}
	return m, nil
}

// Serializes a value to bytes in the given format.
//
// Delegates to [json.Marshal] or [yaml.Marshal] depending on the format.
// Returns [ErrUnsupportedFormat] for unknown formats.
func encodeValue(v any, f Format) ([]byte, error) {
	switch f {
	case JSON:
		return json.Marshal(v)
	case YAML:
		return yaml.Marshal(v)
	default:
		return nil, ErrUnsupportedFormat
	}
}
