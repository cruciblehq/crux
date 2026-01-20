package pack

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/protocol/pkg/archive"
	"github.com/cruciblehq/protocol/pkg/manifest"
)

const (

	// Directory where built artifacts are placed (same as build package)
	Dist = "dist"

	// The path of the manifest file within a Crucible resource project.
	Manifestfile = "crucible.yaml"

	// Default output archive name
	PackageOutput = "package.tar.zst"
)

// Packages a built resource into a distributable archive.
//
// Creates a zstd-compressed tar archive containing the manifest and build
// artifacts from the dist/ directory.
func Pack(ctx context.Context) error {

	// Load manifest to determine resource type
	man, err := manifest.Read(Manifestfile)
	if err != nil {
		return err
	}

	// Check if dist/ exists by attempting to read it
	if _, err := os.ReadDir(Dist); err != nil {
		if os.IsNotExist(err) {
			return crex.UserError("build artifacts not found", "dist/ directory does not exist").
				Fallback("Run 'crux build' first to generate the distribution artifacts.").
				Err()
		}
		return crex.Wrap(ErrFileSystemOperation, err)
	}

	// Validate resource structure based on type
	mismatch := crex.ProgrammingError("pack failed", "manifest config type mismatch").
		Fallback("Please report this issue to the Crucible team.")

	switch man.Resource.Type {
	case "widget":
		widget, ok := man.Config.(*manifest.Widget)
		if !ok {
			return mismatch.Err()
		}
		if err := archive.ValidateWidgetStructure(Dist, widget); err != nil {
			return crex.UserError("widget build output not found", "dist/index.js does not exist").
				Fallback("Run 'crux build' to generate the widget bundle.").
				Err()
		}

	case "service":
		service, ok := man.Config.(*manifest.Service)
		if !ok {
			return mismatch.Err()
		}
		if err := archive.ValidateServiceStructure(Dist, service); err != nil {
			return crex.UserError("service build output not found", "dist/image.tar does not exist").
				Fallback("Run 'crux build' to prepare the service image.").
				Err()
		}

	default:
		return ErrInvalidResourceType
	}

	// Create archive
	return createArchive(PackageOutput)
}

// Creates the archive with manifest and dist/ contents.
func createArchive(outputPath string) error {

	// Create temporary directory for packaging
	tmpDir, err := os.MkdirTemp("", "crux-pack-*")
	if err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy manifest
	manifestDest := filepath.Join(tmpDir, Manifestfile)
	if err := copyFile(Manifestfile, manifestDest); err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}

	// Copy dist/ directory
	distDest := filepath.Join(tmpDir, Dist)
	if err := copyDir(Dist, distDest); err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}

	// Create archive
	if err := archive.Create(tmpDir, outputPath); err != nil {
		return crex.UserError("failed to create package archive", err.Error()).
			Fallback("Check that you have write permissions for the output path.").
			Err()
	}

	return nil
}

// Copies a file from src to dst.
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// Copies a directory recursively.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFile(path, destPath)
	})
}
