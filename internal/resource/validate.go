package resource

import (
	"os"
	"path/filepath"

	"github.com/cruciblehq/spec/manifest"
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
	if _, err := os.Stat(imagePath); err != nil {
		if os.IsNotExist(err) {
			return ErrInvalidStructure
		}
		return err
	}
	return nil
}

// Checks that a widget's build/ directory contains required files.
func validateWidgetStructure(distDir string, _ *manifest.Widget) error {
	widgetMain := filepath.Join(distDir, WidgetMainFile)
	if _, err := os.Stat(widgetMain); err != nil {
		if os.IsNotExist(err) {
			return ErrInvalidStructure
		}
		return err
	}
	return nil
}

// Checks that the package archive exists at the given path.
func validatePackage(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return ErrInvalidStructure
		}
		return err
	}
	return nil
}
