package manifest

// Represents the type of a Crucible resource.
type ResourceType string

const (
	TypeRuntime    ResourceType = "runtime"    // Runtime resource type.
	TypeService    ResourceType = "service"    // Service resource type.
	TypeTemplate   ResourceType = "template"   // Template resource type.
	TypeWidget     ResourceType = "widget"     // Widget resource type.
	TypeAffordance ResourceType = "affordance" // Affordance resource type.
	TypeBlueprint  ResourceType = "blueprint"  // Blueprint resource type.
)

// Converts a string to a resource type, returning an error if invalid.
func ParseResourceType(s string) (ResourceType, error) {
	switch ResourceType(s) {
	case TypeRuntime, TypeService, TypeTemplate, TypeWidget, TypeAffordance, TypeBlueprint:
		return ResourceType(s), nil
	default:
		return "", ErrInvalidResourceType
	}
}
