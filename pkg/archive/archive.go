package archive

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/pkg/paths"
	"github.com/klauspost/compress/zstd"
)

// Creates a zstd-compressed tar archive from a directory.
//
// The archive contains all files and directories under srcDir with paths
// stored relative to srcDir. Paths in the archive use forward slashes
// regardless of the host operating system.
//
// Only regular files and directories are allowed. Symlinks will cause the
// function to return [ErrSymlink]. Other special file types such as devices
// and sockets will cause the function to return [ErrUnsupportedFileType].
//
// If creation fails, the partially written archive is removed.
func Create(srcDir, destPath string) (err error) {
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateFailed, err)
	}
	defer func() {
		file.Close()
		if err != nil {
			os.Remove(destPath)
		}
	}()

	zw, err := zstd.NewWriter(file)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateFailed, err)
	}
	defer zw.Close()

	tw := tar.NewWriter(zw)
	defer tw.Close()

	if err = writeTar(tw, srcDir); err != nil {
		return fmt.Errorf("%w: %w", ErrCreateFailed, err)
	}

	return nil
}

// Extracts a zstd-compressed tar archive to a directory.
//
// Files are extracted with [paths.DefaultFileMode] and directories with
// [paths.DefaultDirMode]. Returns [ErrDestinationExists] if destDir already
// exists.
//
// Only regular files and directories are allowed. Symlinks will cause the
// function to return [ErrSymlink]. Other special file types will cause the
// function to return [ErrUnsupportedFileType]. Path traversal attacks are
// prevented by rejecting paths containing ".." components or absolute paths.
//
// If extraction fails, the destination directory and its contents are removed.
func Extract(archivePath, destDir string) (err error) {
	if _, statErr := os.Stat(destDir); statErr == nil {
		return fmt.Errorf("%w: %s", ErrDestinationExists, destDir)
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrExtractFailed, err)
	}
	defer file.Close()

	zr, err := zstd.NewReader(file)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrExtractFailed, err)
	}
	defer zr.Close()

	if err = os.MkdirAll(destDir, paths.DefaultDirMode); err != nil {
		return fmt.Errorf("%w: %w", ErrExtractFailed, err)
	}
	defer func() {
		if err != nil {
			os.RemoveAll(destDir)
		}
	}()

	if err = readTar(tar.NewReader(zr), destDir); err != nil {
		return fmt.Errorf("%w: %w", ErrExtractFailed, err)
	}

	return nil
}

// Extracts a zstd-compressed tar archive from a reader to a directory.
//
// Same behavior as [Extract] but reads from an [io.Reader] instead of a file.
func ExtractReader(r io.Reader, destDir string) (err error) {
	if _, statErr := os.Stat(destDir); statErr == nil {
		return fmt.Errorf("%w: %s", ErrDestinationExists, destDir)
	}

	zr, err := zstd.NewReader(r)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrExtractFailed, err)
	}
	defer zr.Close()

	if err = os.MkdirAll(destDir, paths.DefaultDirMode); err != nil {
		return fmt.Errorf("%w: %w", ErrExtractFailed, err)
	}
	defer func() {
		if err != nil {
			os.RemoveAll(destDir)
		}
	}()

	if err = readTar(tar.NewReader(zr), destDir); err != nil {
		return fmt.Errorf("%w: %w", ErrExtractFailed, err)
	}

	return nil
}

// Writes directory contents to a tar writer.
func writeTar(tw *tar.Writer, srcDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		return writeEntry(tw, path, relPath, d)
	})
}

// Writes a single entry to the tar writer.
func writeEntry(tw *tar.Writer, path, relPath string, d fs.DirEntry) error {

	info, err := d.Info()
	if err != nil {
		return err
	}

	mode := info.Mode()

	if mode&os.ModeSymlink != 0 {
		return fmt.Errorf("%w: %s", ErrSymlink, relPath)
	}

	if !mode.IsRegular() && !mode.IsDir() {
		return fmt.Errorf("%w: %s", ErrUnsupportedFileType, relPath)
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}

	// Override name and mode
	header.Name = filepath.ToSlash(relPath)
	header.Mode = int64(paths.DefaultFileMode)
	if info.IsDir() {
		header.Mode = int64(paths.DefaultDirMode)
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if mode.IsRegular() {
		return copyFile(tw, path)
	}

	return nil
}

// Copies file contents to a writer.
func copyFile(w io.Writer, path string) error {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}

// Reads tar entries and extracts them to destDir.
func readTar(tr *tar.Reader, destDir string) error {
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		target, err := safePath(destDir, header.Name)
		if err != nil {
			return err
		}

		if err := extractEntry(header, tr, target); err != nil {
			return err
		}
	}
}

// Returns a safe target path within destDir.
func safePath(destDir, name string) (string, error) {

	localName, err := filepath.Localize(name)
	if err != nil {
		return "", fmt.Errorf("%w: %q", ErrInvalidPath, name)
	}

	return filepath.Join(destDir, localName), nil
}

// Extracts a single tar entry.
func extractEntry(header *tar.Header, tr *tar.Reader, target string) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, paths.DefaultDirMode)

	case tar.TypeReg:
		return extractFile(tr, target)

	case tar.TypeSymlink:
		return fmt.Errorf("%w: %s", ErrSymlink, header.Name)

	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedFileType, header.Name)
	}
}

// Extracts a regular file.
func extractFile(r io.Reader, target string) error {

	if err := os.MkdirAll(filepath.Dir(target), paths.DefaultDirMode); err != nil {
		return err
	}

	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, paths.DefaultFileMode)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}
