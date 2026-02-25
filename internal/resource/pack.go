package resource

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/spec/archive"
	"github.com/cruciblehq/spec/manifest"
)

// Holds the output of a successful [Runner.Pack] call.
type PackResult struct {
	Output string // Path where the package archive was written.
}

// Packages a built resource into a distributable archive.
//
// Validates the dist directory and resource structure, then creates a
// zstd-compressed tar archive containing the manifest file and build
// artifacts.
func pack(_ context.Context, m manifest.Manifest, manifestPath, dist, output string) (*PackResult, error) {
	if err := verifyDist(dist); err != nil {
		return nil, err
	}

	if err := verifyResourceStructure(&m, dist); err != nil {
		return nil, err
	}

	if err := ensureOutputDir(output); err != nil {
		return nil, err
	}

	if err := createArchive(output, manifestPath, dist); err != nil {
		return nil, err
	}

	return &PackResult{Output: output}, nil
}

// Verifies whether the build/ directory exists.
func verifyDist(dist string) error {
	if _, err := os.ReadDir(dist); err != nil {
		if os.IsNotExist(err) {
			return crex.UserError("build artifacts not found", "build/ directory does not exist").
				Fallback("Run 'crux build' first to generate the build artifacts.").
				Err()
		}
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	return nil
}

// Verifies the resource structure based on type.
func verifyResourceStructure(man *manifest.Manifest, dist string) error {
	mismatch := crex.ProgrammingError("pack failed", "manifest config type mismatch").
		Fallback("Please report this issue to the Crucible team.")

	switch man.Resource.Type {
	case manifest.TypeRuntime:
		if _, ok := man.Config.(*manifest.Runtime); !ok {
			return mismatch.Err()
		}
		if err := validateImageStructure(dist); err != nil {
			return crex.UserError("runtime build output not found", "build/image.tar does not exist").
				Fallback("Run 'crux build' to prepare the runtime image.").
				Cause(err).
				Err()
		}

	case manifest.TypeService:
		if _, ok := man.Config.(*manifest.Service); !ok {
			return mismatch.Err()
		}
		if err := validateImageStructure(dist); err != nil {
			return crex.UserError("service build output not found", "build/image.tar does not exist").
				Fallback("Run 'crux build' to prepare the service image.").
				Cause(err).
				Err()
		}

	case manifest.TypeWidget:
		widget, ok := man.Config.(*manifest.Widget)
		if !ok {
			return mismatch.Err()
		}
		if err := validateWidgetStructure(dist, widget); err != nil {
			return crex.UserError("widget build output not found", "build/index.js does not exist").
				Fallback("Run 'crux build' to generate the widget bundle.").
				Cause(err).
				Err()
		}

	default:
		return crex.Wrapf(ErrInvalidResourceType, "resource type %q is not supported", man.Resource.Type)
	}

	return nil
}

// Ensures the output directory exists for the package archive.
func ensureOutputDir(outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	if outputDir == "." || outputDir == "" {
		return nil
	}
	if err := os.MkdirAll(outputDir, paths.DefaultDirMode); err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	return nil
}

// Creates the archive with manifest and build/ contents.
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

	// Copy dist/ contents into the archive root
	if err := copyDir(dist, tmpDir); err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}

	// Create archive
	if err := archive.Create(tmpDir, outputPath); err != nil {
		return crex.UserError("failed to create package archive", "could not write the archive to disk").
			Fallback("Check that you have write permissions for the output path.").
			Cause(err).
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

	if _, err := io.Copy(destination, source); err != nil {
		destination.Close()
		return err
	}

	return destination.Close()
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
