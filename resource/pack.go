package resource

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/archive"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/files"
	"github.com/cruciblehq/crux/source"
)

// Holds the output of a successful [Pack] call.
type PackResult struct {
	Output string // Path where the package archive was written.
}

// Packages a built resource into a distributable archive.
//
// Reads the manifest from the build directory and creates a zstd-compressed
// tar archive containing the manifest and build artifacts. The output
// extension must be .tar.zst.
func Pack(_ context.Context, buildDir, output string) (*PackResult, error) {
	if err := ensureOutputDir(output); err != nil {
		return nil, err
	}

	if err := createArchive(output, buildDir); err != nil {
		return nil, err
	}

	return &PackResult{Output: output}, nil
}

// Ensures the output directory exists for the package archive.
func ensureOutputDir(outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	if outputDir == "." || outputDir == "" {
		return nil
	}
	if err := os.MkdirAll(outputDir, files.DefaultDirMode); err != nil {
		return crex.Wrap(source.ErrFileSystemOperation, err)
	}
	return nil
}

// Creates the archive from the build directory contents.
//
// Archives the build directory as-is. The manifest and build artifacts are
// assumed to already be in their final form.
func createArchive(outputPath string, buildDir string) error {
	if err := archive.Create(buildDir, outputPath); err != nil {
		return crex.UserError("failed to create package archive", "could not write the archive to disk").
			Fallback("Check that you have write permissions for the output path.").
			Cause(err).
			Err()
	}

	return nil
}
