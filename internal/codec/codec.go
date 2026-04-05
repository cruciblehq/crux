package codec

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

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
// Field names are matched by the codec struct tag. Type coercions
// such as string-to-int are applied automatically. Fields with a
// "default=X" tag option receive that default when absent from the
// input. Optional decode hooks run before the built-in hooks.
func FromMap(m any, v any, hooks ...mapstructure.DecodeHookFunc) error {
	allHooks := append(hooks, mapstructure.TextUnmarshallerHookFunc())
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          Tag,
		Result:           v,
		Squash:           true,
		WeaklyTypedInput: true,
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(allHooks...),
	})
	if err != nil {
		return err
	}
	if err := d.Decode(m); err != nil {
		return err
	}
	src, _ := m.(map[string]any)
	return applyDefaults(reflect.ValueOf(v), src)
}

// Applies tag-declared defaults to unset fields in a struct.
//
// Walks v recursively through struct and pointer-to-struct fields. A default
// is applied only when the field's key is absent from the source map src and
// the field holds its Go zero value. Fields that were explicitly provided in
// the input (even as zero) are never overwritten. When src is nil the function
// falls back to zero-value detection.
func applyDefaults(v reflect.Value, src map[string]any) error {
	v = deref(v)
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		if err := applyFieldDefault(v.Field(i), t.Field(i), src); err != nil {
			return err
		}
	}
	return nil
}

// Applies the default for a single struct field.
//
// If the field is a struct or pointer-to-struct it recurses via [applyDefaults]
// with the corresponding nested map. Otherwise the default is applied only when
// the field's key is absent from src and the field holds its Go zero value.
// Explicitly provided values are never overwritten.
func applyFieldDefault(field reflect.Value, sf reflect.StructField, src map[string]any) error {
	tag := sf.Tag.Get(Tag)
	name, _, _ := strings.Cut(tag, ",")

	if recurseInto(field) {
		return applyDefaults(field, nestedSource(src, name))
	}

	if src != nil {
		if _, present := src[name]; present {
			return nil
		}
	}

	if !field.IsZero() {
		return nil
	}

	def, ok := parseDefault(tag)
	if !ok {
		return nil
	}
	if err := setDefault(field, def); err != nil {
		return fmt.Errorf("field %s: %w", sf.Name, err)
	}
	return nil
}

// Returns the nested source map for a struct field.
//
// For squashed fields (empty name) the parent map is returned so that key
// lookups stay at the same level. For named fields the corresponding nested
// map is extracted. Returns nil when the key is absent or not a map.
func nestedSource(src map[string]any, name string) map[string]any {
	if src == nil {
		return nil
	}
	if name == "" {
		return src
	}
	m, _ := src[name].(map[string]any)
	return m
}

// Reports whether a field should be recursed into for defaults.
//
// Returns true for struct values and for non-nil pointers to structs. All
// other kinds return false.
func recurseInto(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.Struct:
		return true
	case reflect.Ptr:
		return !field.IsNil() && field.Type().Elem().Kind() == reflect.Struct
	}
	return false
}

// Dereferences a pointer Value to its element.
//
// Returns the element for non-nil pointers, the zero [reflect.Value]
// for nil pointers, and the value unchanged for non-pointer kinds.
func deref(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		return v.Elem()
	}
	return v
}

// Extracts the default value from a codec struct tag.
//
// Splits the tag on commas and looks for an option prefixed with
// "default=". Returns the value and true if found, or an empty
// string and false otherwise.
func parseDefault(tag string) (string, bool) {
	for _, opt := range strings.Split(tag, ",")[1:] {
		if v, ok := strings.CutPrefix(opt, "default="); ok {
			return v, true
		}
	}
	return "", false
}

// Assigns a string default to a scalar reflect.Value.
//
// Parses val into the field's kind using the appropriate strconv
// function. Supports string, bool, all int/uint widths, and
// float32/float64. Returns an error for unsupported kinds or if
// the string cannot be parsed.
func setDefault(field reflect.Value, val string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(val)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 0, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(val, 0, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(n)
	default:
		return fmt.Errorf("unsupported type %s", field.Type())
	}
	return nil
}

// Serializes a map to bytes in the given format.
//
// Delegates to [json.Marshal] or [yaml.Marshal] depending on the
// format. Returns [ErrUnsupportedFormat] for unknown formats.
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

// Deserializes bytes into a map from the given format.
//
// Delegates to [json.Unmarshal] or [yaml.Unmarshal] depending on
// the format. Returns [ErrUnsupportedFormat] for unknown formats.
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
