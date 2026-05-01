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

// Whether any tagged scalar (or scalar within a tagged struct) in cur differs
// from the value at the same position in def.
//
// Walks the struct recursively. Untagged struct fields are descended into;
// tagged struct fields are descended only when the type is not cheaply
// comparable. Tagged slice fields count as non-default when they carry any
// element. Anything else is compared via reflect.Value.Equal.
func hasNonDefaultScalar(t reflect.Type, cur, def reflect.Value) bool {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get(knobStructTag) != "" {
			if taggedFieldNonDefault(f, cur.Field(i), def.Field(i)) {
				return true
			}
			continue
		}
		if f.Type.Kind() == reflect.Struct && hasNonDefaultScalar(f.Type, cur.Field(i), def.Field(i)) {
			return true
		}
	}
	return false
}

// Whether the tagged field at fv differs from its default at dfv.
//
// Composite struct fields are compared field-by-field (or by Equal if cheaply
// comparable); slice fields are non-default when non-empty; scalars are compared
// via reflect.Value.Equal.
func taggedFieldNonDefault(f reflect.StructField, fv, dfv reflect.Value) bool {
	switch {
	case f.Type.Kind() == reflect.Struct:
		if f.Type.Comparable() {
			return !fv.Equal(dfv)
		}
		return hasNonDefaultScalar(f.Type, fv, dfv)
	case f.Type.Kind() == reflect.Slice:
		return fv.Len() > 0
	case f.Type.Comparable():
		return !fv.Equal(dfv)
	}
	return false
}
