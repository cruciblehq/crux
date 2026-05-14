package manifest

import (
	"github.com/cruciblehq/crux/crex"
)

// Widget resources are frontend components that can be embedded into apps.
//
// This structure defines configurations that are unique to widget resource
// manifests. It is used as [Manifest.Config] for [ResourceType.Widget].
type Widget struct {

	// Declared parameters for this widget.
	//
	// Lists configuration values the widget accepts when embedded. Values
	// are bound through environment declarations.
	Schema *Schema `codec:"schema,omitempty"`

	// Build entry point.
	//
	// A path to the widget's main file, relative to the manifest. This is the
	// file that will be built and bundled for distribution. The build output
	// is expected to be a single JavaScript file.
	Main string `codec:"main"`
}

// Validates the widget configuration.
func (w *Widget) Validate() error {
	if w.Main == "" {
		return crex.Wrap(ErrInvalidWidget, ErrMissingMain)
	}

	if w.Schema != nil {
		if err := w.Schema.Validate(); err != nil {
			return crex.Wrap(ErrInvalidWidget, err)
		}
	}

	return nil
}
