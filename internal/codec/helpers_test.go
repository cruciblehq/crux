package codec

type sample struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type nested struct {
	Inner sample `json:"inner"`
}

type squashed struct {
	sample `json:",squash"`
	Extra  string `json:"extra"`
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
