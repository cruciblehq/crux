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

// Options for packaging a Crucible resource.
type Options struct {
	Manifest string // Path to the manifest file.
	Dist     string // Directory where built artifacts are located.
	Output   string // Output archive path.
}

// Result of packaging a Crucible resource.
type Result struct {
	Output string // Path where the package was written.
}

// Packages a built resource into a distributable archive.
//
// Creates a zstd-compressed tar archive containing the manifest and build
// artifacts from the directory specified by opts.Dist.
func Pack(ctx context.Context, opts Options) (*Result, error) {

	// Load manifest to determine resource type
	man, err := manifest.Read(opts.Manifest)
	if err != nil {
		return nil, err
	}

	// Check if dist/ exists by attempting to read it
	if _, err := os.ReadDir(opts.Dist); err != nil {
		if os.IsNotExist(err) {
			return nil, crex.UserError("build artifacts not found", "dist/ directory does not exist").
				Fallback("Run 'crux build' first to generate the distribution artifacts.").
				Err()
		}
		return nil, crex.Wrap(ErrFileSystemOperation, err)
	}

	// Validate resource structure based on type
	mismatch := crex.ProgrammingError("pack failed", "manifest config type mismatch").
		Fallback("Please report this issue to the Crucible team.")

	switch man.Resource.Type {
	case "widget":
		widget, ok := man.Config.(*manifest.Widget)
		if !ok {
			return nil, mismatch.Err()
		}
		if err := archive.ValidateWidgetStructure(opts.Dist, widget); err != nil {
			return nil, crex.UserError("widget build output not found", "dist/index.js does not exist").
				Fallback("Run 'crux build' to generate the widget bundle.").
				Err()
		}

	case "service":
		service, ok := man.Config.(*manifest.Service)
		if !ok {
			return nil, mismatch.Err()
		}
		if err := archive.ValidateServiceStructure(opts.Dist, service); err != nil {
			return nil, crex.UserError("service build output not found", "dist/image.tar does not exist").
				Fallback("Run 'crux build' to prepare the service image.").
				Err()
		}

	default:
		return nil, ErrInvalidResourceType
	}

	if err := createArchive(opts.Output, opts.Manifest, opts.Dist); err != nil {
		return nil, err
	}

	return &Result{
		Output: opts.Output,
	}, nil
}

// Creates the archive with manifest and dist/ contents.
func createArchive(outputPath, manifestfile, dist string) error {

	// Create temporary directory for packaging
	tmpDir, err := os.MkdirTemp("", "crux-pack-*")
	if err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy manifest
	manifestDest := filepath.Join(tmpDir, filepath.Base(manifestfile))
	if err := copyFile(manifestfile, manifestDest); err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}

	// Copy dist/ directory
	distDest := filepath.Join(tmpDir, filepath.Base(dist))
	if err := copyDir(dist, distDest); err != nil {
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
