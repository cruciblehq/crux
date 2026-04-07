package codec

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Applies tag-declared defaults to unset fields in a struct.
//
// Walks v recursively through struct and pointer-to-struct fields. A default
// is applied only when the field's key is absent from the source map src.
// Fields that were explicitly provided in the input (even as zero) are never
// overwritten.
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
// with a nested map. Otherwise the default is applied only when the field's key
// is absent from src. Values provided explicitly are never overwritten.
func applyFieldDefault(field reflect.Value, sf reflect.StructField, src map[string]any) error {
	rawTag := sf.Tag.Get(tag)
	name, _, _ := strings.Cut(rawTag, ",")

	if recurseInto(field) {
		return applyDefaults(field, nestedSource(src, name))
	}

	if src != nil {
		if _, present := src[name]; present {
			return nil
		}
	}

	def, ok := parseDefault(rawTag)
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
// Splits the tag on commas and looks for an option prefixed with "default=".
// Returns the value and true if found, or an empty string and false otherwise.
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
