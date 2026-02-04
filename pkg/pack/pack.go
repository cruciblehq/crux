package pack

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/cruciblehq/protocol/pkg/archive"
	"github.com/cruciblehq/protocol/pkg/crex"
	"github.com/cruciblehq/protocol/pkg/manifest"
	"github.com/cruciblehq/protocol/pkg/resource"
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
	man, err := manifest.Read(opts.Manifest)
	if err != nil {
		return nil, err
	}

	if err := validateDistExists(opts.Dist); err != nil {
		return nil, err
	}

	if err := validateResourceStructure(man, opts.Dist); err != nil {
		return nil, err
	}

	if err := createArchive(opts.Output, opts.Manifest, opts.Dist); err != nil {
		return nil, err
	}

	return &Result{Output: opts.Output}, nil
}

// Validates that the dist/ directory exists.
func validateDistExists(dist string) error {
	if _, err := os.ReadDir(dist); err != nil {
		if os.IsNotExist(err) {
			return crex.UserError("build artifacts not found", "dist/ directory does not exist").
				Fallback("Run 'crux build' first to generate the distribution artifacts.").
				Err()
		}
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	return nil
}

// Validates the resource structure based on type.
func validateResourceStructure(man *manifest.Manifest, dist string) error {
	mismatch := crex.ProgrammingError("pack failed", "manifest config type mismatch").
		Fallback("Please report this issue to the Crucible team.")

	switch resource.Type(man.Resource.Type) {
	case resource.TypeRuntime:
		if _, ok := man.Config.(*manifest.Runtime); !ok {
			return mismatch.Err()
		}
		if err := validateImageStructure(dist); err != nil {
			return crex.UserError("runtime build output not found", "dist/image.tar does not exist").
				Fallback("Run 'crux build' to generate the runtime image.").
				Cause(err).
				Err()
		}

	case resource.TypeService:
		if _, ok := man.Config.(*manifest.Service); !ok {
			return mismatch.Err()
		}
		if err := validateImageStructure(dist); err != nil {
			return crex.UserError("service build output not found", "dist/image.tar does not exist").
				Fallback("Run 'crux build' to prepare the service image.").
				Cause(err).
				Err()
		}

	case resource.TypeWidget:
		widget, ok := man.Config.(*manifest.Widget)
		if !ok {
			return mismatch.Err()
		}
		if err := validateWidgetStructure(dist, widget); err != nil {
			return crex.UserError("widget build output not found", "dist/index.js does not exist").
				Fallback("Run 'crux build' to generate the widget bundle.").
				Cause(err).
				Err()
		}

	default:
		return ErrInvalidResourceType
	}

	return nil
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
