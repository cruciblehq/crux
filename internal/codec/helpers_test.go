package codec

type sample struct {
	Name    string `codec:"name"`
	Version int    `codec:"version"`
}

type nested struct {
	Inner sample `codec:"inner"`
}

type squashed struct {
	sample `codec:",squash"`
	Extra  string `codec:"extra"`
}

type custom struct {
	Value string
}

func (c *custom) Encode() (any, error) {
	return map[string]any{"custom": c.Value}, nil
}

func (c *custom) Decode(raw any) error {
	m, _ := raw.(map[string]any)
	c.Value, _ = m["custom"].(string)
	return nil
}
