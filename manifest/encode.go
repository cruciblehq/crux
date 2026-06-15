package manifest

import (
	"maps"

	"github.com/cruciblehq/crux/codec"
	"github.com/cruciblehq/crux/crex"
)

// Converts v to a map[string]any.
//
// If v implements [codec.Encodable], its Encode method is called and the
// result is asserted as a map[string]any. Otherwise [codec.Codec.ToMap] is
// used. Returns [ErrEncodeFailed] if the encoded value is not a map.
func encodeToMap(c *codec.Codec, v any) (map[string]any, error) {
	if enc, ok := v.(codec.Encodable); ok {
		raw, err := enc.Encode(c)
		if err != nil {
			return nil, err
		}
		m, ok := raw.(map[string]any)
		if !ok {
			return nil, crex.Wrapf(ErrEncodeFailed, "encoded value %T is not a map", raw)
		}
		return m, nil
	}
	return c.ToMap(v)
}

// Merges src into dst, returning an error if any key appears in both.
//
// On success dst is modified in place and returned. The merge is safe to use
// for combining two disjoint encoded maps (for example, a manifest envelope
// and a config block) where key collisions indicate a structural problem.
func mergeMap(dst, src map[string]any) (map[string]any, error) {
	for k := range dst {
		if _, exists := src[k]; exists {
			return nil, crex.Wrapf(ErrEncodeFailed, "key %q conflicts", k)
		}
	}
	maps.Copy(dst, src)
	return dst, nil
}
