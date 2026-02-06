package pack

import (
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/manifest"
)

const (

	// The required OCI image file for runtimes and services.
	ImageFile = "image.tar"

	// The required main file for widgets.
	WidgetMainFile = "index.js"
)

// Checks that an image-based resource's build/ directory contains the image.
func validateImageStructure(distDir string) error {
	imagePath := filepath.Join(distDir, ImageFile)
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return ErrInvalidStructure
	}
	return nil
}

// Checks that a widget's build/ directory contains required files.
func validateWidgetStructure(distDir string, m *manifest.Widget) error {
	widgetMain := filepath.Join(distDir, WidgetMainFile)
	if _, err := os.Stat(widgetMain); os.IsNotExist(err) {
		return ErrInvalidStructure
	}
	return nil
}
