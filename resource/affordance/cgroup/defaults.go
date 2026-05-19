package cgroup

import (
	"reflect"
	"strconv"
)

// Tag used to specify default values for struct fields.
//
// The tag value is parsed according to the field's kind and set on the field
// during spec construction. Only scalar kinds are supported; struct fields
// must be tagged at the leaf level. The tag is ignored on unsupported kinds.
const defaultTag = "default"

// Walks struct type t and writes each field's `default:` tag value into v.
//
// Recurses into embedded struct fields. Fields without a default tag and
// fields of non-scalar kinds are left at their Go zero values.
func setDefaults(t reflect.Type, v reflect.Value) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)
		if def := f.Tag.Get(defaultTag); def != "" {
			setDefault(fv, def)
			continue
		}
		if f.Type.Kind() == reflect.Struct {
			setDefaults(f.Type, fv)
		}
	}
}

// Sets a field to its default value from a struct tag string.
//
// Dispatches on reflect.Kind so that no knob catalog is needed. Covers
// bool, unsigned integers, signed integers, and string-kinded types
// (which includes string-based typedefs such as nodeType and partition).
func setDefault(field reflect.Value, def string) {
	switch field.Kind() {
	case reflect.Bool:
		v, _ := strconv.ParseBool(def)
		field.SetBool(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, _ := strconv.ParseUint(def, 10, field.Type().Bits())
		field.SetUint(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, _ := strconv.ParseInt(def, 10, field.Type().Bits())
		field.SetInt(v)
	case reflect.String:
		field.SetString(def)
	default:
		panic("cgroup: unsupported kind for default tag: " + field.Kind().String())
	}
}
