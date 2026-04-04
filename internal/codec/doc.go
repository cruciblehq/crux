// Package codec provides format-neutral encoding and decoding.
//
// All types in the spec module carry a single codec:"name" struct tag. The
// codec package converts between Go values and byte representations in JSON
// or YAML using mapstructure as the struct-to-map bridge and standard library
// encoders for the final byte format.
//
//	data, err := codec.Encode(v, codec.JSON)
//	err := codec.Decode(data, &v, codec.YAML)
package codec
