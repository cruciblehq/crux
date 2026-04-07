// Package codec provides format-neutral encoding and decoding.
//
// All types in the project carry a single codec:"name" struct tag. The codec
// package converts between Go values and byte representations in JSON or YAML
// using mapstructure as the struct-to-map bridge and standard library encoders
// for the final byte format.
//
// [Encode] serializes a Go value to bytes. If the value implements [Encodable],
// its Encode method produces the intermediate representation; otherwise [ToMap]
// converts the struct to a map using struct tags. The result is then marshaled
// to the requested format.
//
//	data, err := codec.Encode(v, codec.YAML, "codec")
//
// [Unmarshal] parses bytes (JSON or YAML) into a map, then delegates to
// [Decode] to populate a Go struct. [Decode] walks the type tree, applying
// weak type coercion (e.g. string-to-int) and tag-declared defaults for
// absent fields.
//
//	err := codec.Unmarshal(data, &v, codec.YAML, "codec")
//
// Types that need custom decoding logic implement [Decodable]. When [Decode]
// encounters such a type at any depth, it calls Decode with the raw
// map[string]any instead of mapping fields automatically. Implementations
// use [Field] to decode individual fields one at a time, retaining coercion
// and defaults, and call [Decode] on nested structs of different types.
//
//	func (m *MyType) Decode(raw map[string]any) error {
//	    codec.Field(raw, m, "Name", "codec")
//	    codec.Field(raw, m, "Version", "codec")
//	    // custom dispatch or iteration for remaining fields
//	    return nil
//	}
//
// Struct fields may declare a default value in their tag:
//
//	Weight uint16 `codec:"weight,default=100"`
//
// Defaults are applied only when the field's key is absent from the source
// map. Explicitly provided values, including zero, are never overwritten.
//
// [Validate] calls the [Validatable.Validate] method on a value. Validation
// is always a separate step from decoding, allowing partial or invalid data
// to be decoded first and validated later.
package codec
