package resource

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/utils-go/archive"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Holds the output of a successful [Pack] call.
type PackResult struct {
	Output string // Path where the package archive was written.
}

// Packages a built resource into a distributable archive.
//
// Archives the build directory, which holds the manifest and the build
// artifacts, into output. The compression follows the output extension, which
// must be .tar.zst.
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
	if err := os.MkdirAll(outputDir, file.DefaultDirMode); err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	return nil
}

// Creates the archive from the build directory contents.
//
// Archives the build directory unchanged. The manifest and build artifacts are
// assumed to already be in their final form. The compression follows the
// output extension, and the archive appears at outputPath only once it is
// complete.
func createArchive(outputPath, buildDir string) error {
	const (
		description = "failed to create package archive"
		reason      = "could not write the archive to disk"
		recovery    = "Check that you have write permissions for the output path."
	)

	format, err := archive.FormatOfName(outputPath)
	if err != nil {
		return crex.UserError(description, "the output path does not name a supported archive format").
			Recovery("Use an output path that ends in .tar.zst.").
			Cause(err).
			Err()
	}

	af, err := file.CreateAtomic(filepath.Dir(outputPath), file.DefaultFileMode)
	if err != nil {
		return crex.SystemError(description, reason).
			Recovery(recovery).
			Cause(crex.Wrap(ErrFileSystemOperation, err)).
			Err()
	}
	defer af.Abort()

	if err := archive.Write(af, buildDir, "", format); err != nil {
		return crex.SystemError(description, reason).
			Recovery(recovery).
			Cause(err).
			Err()
	}
	if err := af.Commit(outputPath); err != nil {
		return crex.SystemError(description, reason).
			Recovery(recovery).
			Cause(crex.Wrap(ErrFileSystemOperation, err)).
			Err()
	}

	return nil
}
