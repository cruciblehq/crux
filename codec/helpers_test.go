package codec

// Test fixture with a name and a version number.
type sample struct {
	Name    string `codec:"name"`    // Name string.
	Version int    `codec:"version"` // Version number.
}

// Test fixture with a named inner sample.
type nested struct {
	Inner sample `codec:"inner"` // Nested sample value.
}

// Test fixture that squashes a sample into the parent map.
type squashed struct {
	sample        // Embedded name and version fields.
	Extra  string `codec:"extra"` // Additional field alongside the squashed fields.
}

// Test fixture that implements [Encodable] and [Decodable].
type custom struct {
	Value string // Encoded or decoded string value.
}

// Encodes custom to a map under the "custom" key.
func (c *custom) Encode(_ *Codec) (any, error) {
	return map[string]any{"custom": c.Value}, nil
}

// Decodes the "custom" key from a raw map into Value.
func (c *custom) Decode(_ *Codec, raw any) error {
	m, _ := raw.(map[string]any)
	c.Value, _ = m["custom"].(string)
	return nil
}
