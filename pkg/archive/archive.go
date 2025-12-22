package archive

import (
	"archive/tar"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/paths"
	"github.com/klauspost/compress/zstd"
)

var (
	ErrCreateFailed        = errors.New("archive creation failed")
	ErrExtractFailed       = errors.New("extraction failed")
	ErrInvalidPath         = errors.New("invalid path")
	ErrUnsupportedFileType = errors.New("unsupported file type")
)

// Creates a zstd-compressed tar archive from a directory.
//
// The archive contains all files and directories under src with paths stored
// relative to src. Paths in the archive use forward slashes regardless of the
// host operating system.
//
// Only regular files and directories are allowed. Symlinks and other special
// file types such as devices and sockets will cause the function to return
// [ErrUnsupportedFileType].
//
// If creation fails, the partially written archive is removed.
func Create(src, dest string) (err error) {
	file, err := os.Create(dest)
	if err != nil {
		return crex.Wrap(ErrCreateFailed, err)
	}
	defer file.Close()

	zw, err := zstd.NewWriter(file)
	if err != nil {
		os.Remove(dest)
		return crex.Wrap(ErrCreateFailed, err)
	}
	defer func() {
		zw.Close()
		if err != nil {
			os.Remove(dest)
		}
	}()

	tw := tar.NewWriter(zw)
	defer tw.Close()

	if err = writeTar(tw, src); err != nil {
		return crex.Wrap(ErrCreateFailed, err)
	}

	return nil
}

// Extracts a zstd-compressed tar archive to a directory.
//
// Files are extracted with [paths.DefaultFileMode] and directories with
// [paths.DefaultDirMode]. Returns [ErrDestinationExists] if dest already exists.
//
// Only regular files and directories are allowed. Symlinks and other special
// file types return [ErrUnsupportedFileType]. Absolute paths and path traversal
// attempts (e.g., "../etc/passwd") return [ErrInvalidPath].
//
// If extraction fails, the destination directory and its contents are removed.
func Extract(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}
	defer file.Close()

	return ExtractFromReader(file, dest)
}

// ExtractFromReader extracts a zstd-compressed tar archive from a reader to a directory.
//
// Same behavior as [Extract] but reads from an [io.Reader] instead of a file.
func ExtractFromReader(r io.Reader, dest string) error {
	if _, statErr := os.Stat(dest); statErr == nil {
		return crex.Wrap(ErrExtractFailed, os.ErrExist)
	}

	zr, err := zstd.NewReader(r)
	if err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}
	defer zr.Close()

	err = extractToDirectory(tar.NewReader(zr), dest)
	if err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}

	return nil
}

// Extracts tar contents to a directory with proper cleanup on failure.
//
// Creates dest if it doesn't exist. If any error occurs during extraction,
// dest and all extracted contents are removed.
func extractToDirectory(tr *tar.Reader, dest string) (err error) {
	if err = os.MkdirAll(dest, paths.DefaultDirMode); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(dest)
		}
	}()

	if err = readTar(tr, dest); err != nil {
		return err
	}

	return nil
}

// Writes directory contents to a tar writer.
//
// Walks src directory recursively and writes each entry to tw. Paths in the
// archive are relative to src and use forward slashes.
func writeTar(tw *tar.Writer, src string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
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
//
// Validates file type, creates tar header with normalized path and permissions,
// and writes file contents for regular files. Returns [ErrUnsupportedFileType]
// for symlinks and special files.
func writeEntry(tw *tar.Writer, path, relPath string, d fs.DirEntry) error {

	info, err := d.Info()
	if err != nil {
		return err
	}

	mode := info.Mode()

	if mode&os.ModeSymlink != 0 || (!mode.IsRegular() && !mode.IsDir()) {
		return ErrUnsupportedFileType
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

// Copies file contents from path to w.
func copyFile(w io.Writer, path string) error {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}

// Reads tar entries and extracts them to dest.
//
// Validates each entry path for security before extraction. Returns the first
// error encountered or nil on successful completion.
func readTar(tr *tar.Reader, dest string) error {
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		target, err := validateAndJoinPath(dest, header.Name)
		if err != nil {
			return err
		}

		if err := extractEntry(header, tr, target); err != nil {
			return err
		}
	}
}

// Validates and joins an archive path with the destination directory.
//
// Uses [filepath.Localize] to convert slash-separated paths to OS format, then
// [filepath.IsLocal] to ensure the path is local (not absolute, no ".." traversal,
// no reserved names on Windows). Returns the validated path joined with dest.
func validateAndJoinPath(dest, name string) (string, error) {

	localName, err := filepath.Localize(name)
	if err != nil {
		return "", ErrInvalidPath
	}

	// Not empty, not absolute path, no ".." traversal, no reserved names on Windows
	if !filepath.IsLocal(localName) {
		return "", ErrInvalidPath
	}

	return filepath.Join(dest, localName), nil
}

// Extracts a single tar entry to target.
//
// Handles directories and regular files. Returns [ErrUnsupportedFileType]
// for all other entry types including symlinks.
func extractEntry(header *tar.Header, tr *tar.Reader, target string) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return extractDirectory(target)

	case tar.TypeReg:
		return extractFile(tr, target)

	default:
		return ErrUnsupportedFileType
	}
}

// Creates a directory at target with [paths.DefaultDirMode].
//
// Creates all parent directories as needed.
func extractDirectory(target string) error {
	if err := os.MkdirAll(target, paths.DefaultDirMode); err != nil {
		return err
	}
	return nil
}

// Extracts a regular file from r to target.
//
// Creates parent directories as needed, then writes file contents with
// [paths.DefaultFileMode].
func extractFile(r io.Reader, target string) error {

	// Ensure parent directory exists
	if err := extractDirectory(filepath.Dir(target)); err != nil {
		return err
	}

	// Create file
	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, paths.DefaultFileMode)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = io.Copy(f, r); err != nil {
		return err
	}

	return nil
}
