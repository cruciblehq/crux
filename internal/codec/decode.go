package codec

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/cruciblehq/crex"
	"github.com/go-viper/mapstructure/v2"
	"gopkg.in/yaml.v3"
)

// Implemented by types that need custom decoding logic.
//
// When [Decode] encounters a target type that implements this interface
// (at any depth in the struct tree), it delegates to Decode with the raw
// parsed map instead of using the default field-by-field mapping.
type Decodable interface {

	// Decodes a raw parsed map into the receiver.
	//
	// Implementations use [Field] to decode individual fields and [Decode]
	// for nested structs, retaining defaults, coercion, and hook dispatch.
	// Returning an error aborts the outer [Unmarshal] or [Decode] call.
	Decode(raw map[string]any) error
}

// Populates dst from a map.
//
// Field names are matched by the codec struct tag. Type coercions such as
// string-to-int are applied automatically. Fields with a "default=X" tag
// option receive that default when absent from the input. At each node in
// the type tree, if the target implements [Decodable], its Decode method
// is called with the raw map.
func Decode(src map[string]any, dst any) error {
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          tag,
		Result:           dst,
		Squash:           true,
		WeaklyTypedInput: true,
		DecodeHook:       decoderHook(),
	})
	if err != nil {
		return err
	}
	if err := d.Decode(src); err != nil {
		return err
	}
	return applyDefaults(reflect.ValueOf(dst), src)
}

// Populates v from data in the given format.
//
// The data is first parsed into a map, then applied to v via [Decode].
// If v (or any nested field) implements [Decodable], its Decode method
// is called with the raw parsed map at that point in the tree.
func Unmarshal(data []byte, v any, f Format) error {
	m, err := decodeMap(data, f)
	if err != nil {
		return err
	}
	return Decode(m, v)
}

// Decodes a single struct field from a map.
//
// Looks up the field's codec tag key in src. If present, the raw value is
// decoded into the field with type coercion and hook dispatch. If absent,
// the tag-declared default (if any) is applied. Returns a programming error
// if fieldName does not exist on v's type or v is not a pointer to a struct.
func Field(src map[string]any, v any, fieldName string) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return crex.ProgrammingError("decode field failed", "v must be a pointer to a struct").Err()
	}
	rv = rv.Elem()

	sf, ok := rv.Type().FieldByName(fieldName)
	if !ok {
		return crex.ProgrammingErrorf("decode field failed", "%s has no field %q", rv.Type().Name(), fieldName).Err()
	}

	rawTag := sf.Tag.Get(tag)
	key, _, _ := strings.Cut(rawTag, ",")

	rawVal, present := src[key]
	field := rv.FieldByIndex(sf.Index)

	if !present {
		return applyFieldDefault(field, sf, src)
	}

	tmp := reflect.New(sf.Type)
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          tag,
		Result:           tmp.Interface(),
		Squash:           true,
		WeaklyTypedInput: true,
		DecodeHook:       decoderHook(),
	})
	if err != nil {
		return err
	}
	if err := d.Decode(rawVal); err != nil {
		return err
	}
	field.Set(tmp.Elem())

	nested, _ := rawVal.(map[string]any)
	return applyDefaults(field, nested)
}

// Returns a decode hook that delegates to [Decodable.Decode] for any target
// type that implements [Decodable].
func decoderHook() mapstructure.DecodeHookFuncType {
	iface := reflect.TypeOf((*Decodable)(nil)).Elem()
	return func(from, to reflect.Type, data any) (any, error) {
		m, ok := data.(map[string]any)
		if !ok {
			return data, nil
		}
		ptr := reflect.PointerTo(to)
		if !ptr.Implements(iface) {
			return data, nil
		}
		result := reflect.New(to)
		if err := result.Interface().(Decodable).Decode(m); err != nil {
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
