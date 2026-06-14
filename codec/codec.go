package codec

// Encodes and decodes values using a named struct tag.
//
// The tag selects which struct fields map to serialized keys. [Default] uses
// the "codec" tag; [New] binds a different one.
type Codec struct {
	tag string // Struct tag used to map fields.
}

// Returns a Codec bound to the default "codec" struct tag.
func Default() *Codec {
	return &Codec{tag: defaultCodecTag}
}

// Returns a Codec bound to the given struct tag.
//
// An empty tag falls back to the default "codec" tag.
func New(tag string) *Codec {
	if tag == "" {
		tag = defaultCodecTag
	}
	return &Codec{tag: tag}
}

// Converts v to bytes in the given format using the default codec tag.
func Encode(v any, f Format) ([]byte, error) {
	return Default().Encode(v, f)
}

// Populates v from data in the given format using the default codec tag.
func Unmarshal(data []byte, v any, f Format) error {
	return Default().Unmarshal(data, v, f)
}

// Populates dst from a map using the default codec tag.
func Decode(src map[string]any, dst any) error {
	return Default().Decode(src, dst)
}

// Converts a struct to a map using the default codec tag.
func ToMap(v any) (map[string]any, error) {
	return Default().ToMap(v)
}

// Decodes a single struct field from a map using the default codec tag.
func Field(src map[string]any, v any, fieldName string) error {
	return Default().Field(src, v, fieldName)
}
