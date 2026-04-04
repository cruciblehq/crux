package codec

// Serialization format supported by the codec.
type Format int

const (
	JSON Format = iota // JSON serialization format.
	YAML               // YAML serialization format.
)

// Returns the lowercase name of the format.
func (f Format) String() string {
	switch f {
	case JSON:
		return "json"
	case YAML:
		return "yaml"
	default:
		return "unknown"
	}
}
