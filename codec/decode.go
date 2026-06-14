package codec

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/cruciblehq/crux/crex"
	"github.com/go-viper/mapstructure/v2"
	"gopkg.in/yaml.v3"
)

// Implemented by types that need custom decoding logic.
//
// When [Codec.Decode] encounters a target type that implements this interface
// (at any depth in the struct tree), it delegates to Decode with the raw
// value instead of using the default field-by-field mapping. The raw value
// is typically a map[string]any but may be a string or other scalar when
// the source value is not a map.
type Decodable interface {

	// Decodes a raw value into the receiver.
	//
	// The given Codec carries the active tag namespace and should be used for
	// any nested [Codec.Field] or [Codec.Decode] calls. The raw value is the
	// parsed representation from the source format, typically a map[string]any
	// but possibly a string or other scalar. Returning an error aborts the
	// outer [Codec.Unmarshal] or [Codec.Decode] call.
	Decode(c *Codec, raw any) error
}

// Populates dst from a map.
//
// Field names are matched by the codec's struct tag. Type coercions such as
// string-to-int are applied automatically. Fields with a `default:"X"` tag
// receive that default when absent from the input. At each node in
// the type tree, if the target implements [Decodable], its Decode method
// is called with the raw map.
func (c *Codec) Decode(src map[string]any, dst any) error {
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          c.tag,
		Result:           dst,
		Squash:           true,
		WeaklyTypedInput: true,
		DecodeHook:       c.decoderHook(),
	})
	if err != nil {
		return err
	}
	if err := d.Decode(src); err != nil {
		return err
	}
	return c.applyDefaults(reflect.ValueOf(dst), src)
}

// Populates v from data in the given format.
//
// The data is first parsed into a map, then applied to v via [Codec.Decode].
// If v (or any nested field) implements [Decodable], its Decode method is
// called with the raw parsed map at that point in the tree.
func (c *Codec) Unmarshal(data []byte, v any, f Format) error {
	m, err := decodeMap(data, f)
	if err != nil {
		return err
	}
	return c.Decode(m, v)
}

// Decodes a single struct field from a map.
//
// Looks up the field's codec tag key in src. If present, the raw value is
// decoded into the field with type coercion and hook dispatch. If absent,
// the tag-declared default (if any) is applied. Returns a programming error
// if fieldName does not exist on v's type or v is not a pointer to a struct.
func (c *Codec) Field(src map[string]any, v any, fieldName string) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return ErrInvalidInput
	}
	rv = rv.Elem()

	sf, ok := rv.Type().FieldByName(fieldName)
	if !ok {
		return crex.Wrapf(ErrMissingField, "%s has no field %q", rv.Type().Name(), fieldName)
	}

	rawTag := sf.Tag.Get(c.tag)
	key, _, _ := strings.Cut(rawTag, ",")

	rawVal, present := src[key]
	field := rv.FieldByIndex(sf.Index)

	if !present {
		return c.applyFieldDefault(field, sf, src)
	}

	tmp := reflect.New(sf.Type)
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          c.tag,
		Result:           tmp.Interface(),
		Squash:           true,
		WeaklyTypedInput: true,
		DecodeHook:       c.decoderHook(),
	})
	if err != nil {
		return err
	}
	if err := d.Decode(rawVal); err != nil {
		return err
	}
	field.Set(tmp.Elem())

	nested, _ := rawVal.(map[string]any)
	return c.applyDefaults(field, nested)
}

// Returns a decode hook that delegates to [Decodable.Decode] for any target
// type that implements [Decodable].
func (c *Codec) decoderHook() mapstructure.DecodeHookFuncType {
	iface := reflect.TypeFor[Decodable]()
	return func(from, to reflect.Type, data any) (any, error) {
		ptr := reflect.PointerTo(to)
		if !ptr.Implements(iface) {
			return data, nil
		}
		result := reflect.New(to)
		if err := result.Interface().(Decodable).Decode(c, data); err != nil {
			return nil, err
		}
		return result.Elem().Interface(), nil
	}
}

// Deserializes bytes into a map from the given format.
//
// Delegates to [json.Unmarshal] or [yaml.Unmarshal] depending on the format.
// Returns [ErrUnsupportedFormat] for unknown formats.
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
