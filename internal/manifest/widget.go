package manifest

import "github.com/cruciblehq/crex"

// The JavaScript bundle produced by widget builds.
const WidgetMainFile = "index.js"

// Holds configuration specific to widget resources.
//
// Widget resources are frontend components that can be embedded into apps.
// This structure defines configurations that are unique to widget resource
// manifests. It is used as [Manifest.Config] for [ResourceType.Widget].
type Widget struct {

	// Declared parameters for this widget.
	//
	// Lists configuration values the widget accepts when embedded. Values
	// are bound through environment declarations.
	Schema Schema `codec:"schema,omitempty"`

	// Build entry point.
	Main string `codec:"main"`
}

// Validates the widget configuration.
func (w *Widget) Validate() error {
	if w.Main == "" {
		return crex.Wrap(ErrInvalidWidget, ErrMissingMain)
	}

	if err := w.Schema.Validate(); err != nil {
		return crex.Wrap(ErrInvalidWidget, err)
	}

	return nil
}
