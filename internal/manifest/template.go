package manifest

// Holds configuration specific to template resources.
//
// Template resources are reusable resource project structures that can be
// instantiated to create new resources. This structure defines configurations
// that are unique to template resources.
type Template struct {

	// Declared parameters for this template.
	//
	// Lists the values the template accepts when instantiated.
	Schema Schema `codec:"schema,omitempty"`
}

// Validates the template configuration.
func (t *Template) Validate() error {
	return t.Schema.Validate()
}
