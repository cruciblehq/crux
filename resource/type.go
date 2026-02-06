package resource

// Represents the type of a Crucible resource.
type Type string

const (
	TypeService  Type = "service"  // Service resource type.
	TypeTemplate Type = "template" // Template resource type.
	TypeWidget   Type = "widget"   // Widget resource type.
)

// Converts a string to a Type, returning an error if invalid.
func ParseType(s string) (Type, error) {
	switch Type(s) {
	case TypeService, TypeTemplate, TypeWidget:
		return Type(s), nil
	default:
		return "", ErrInvalidType
	}
}
