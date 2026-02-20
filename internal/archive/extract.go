package archive

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cruciblehq/crex"
)

// Extracts a compressed tar archive to a directory.
//
// The compression format is detected from the src filename extension (see
// [Format]). Permissions are preserved from the archive headers with special
// bits stripped. Intermediate directories that have no archive entry are
// created with [os.ModePerm], subject to the process umask. Returns
// [ErrExtractFailed] wrapping [os.ErrExist] if dest already exists. Regular
// files, directories, symlinks, and hard links are supported. Symlinks and
// hard links are validated to ensure they do not escape the destination tree.
// PAX extended headers are skipped transparently. Other entry types such as
// devices and sockets return [ErrUnsupportedFileType]. Absolute paths and path
// traversal attempts (e.g., "../etc/passwd") return [ErrInvalidPath]. If
// extraction fails, the destination directory and its contents are removed.
func Extract(src, dest string) (err error) {
	fmt, err := detect(src)
	if err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}

	if _, statErr := os.Stat(dest); statErr == nil {
		return crex.Wrap(ErrExtractFailed, os.ErrExist)
	}

	file, err := os.Open(src)
	if err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}
	defer file.Close()

	defer func() {
		if err != nil {
			os.RemoveAll(dest)
		}
	}()

	return ExtractFromReader(file, dest, fmt)
}

// Extracts a compressed tar archive from a reader to a directory.
//
// Creates dest if it does not exist and extracts all entries into it. The
// compression format must be supplied explicitly because there is no filename
// to detect from. Permissions are preserved from the archive headers with
// special bits stripped. Intermediate directories that have no archive entry
// are created with [os.ModePerm], subject to the process umask. Supports the
// same entry types as [Extract]: regular files, directories, symlinks, and
// hard links (all validated against directory escape). PAX headers are
// skipped. Unlike [Extract], this function does not check whether dest
// already exists and does not clean up on failure.
func ExtractFromReader(r io.Reader, dest string, f Format) error {
	dr, err := newDecompressReader(r, f)
	if err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}
	defer dr.Close()

	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}

	if err := readTar(tar.NewReader(dr), dest); err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}

	return nil
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

		if target == "" {
			continue
		}

		if err := extractEntry(header, tr, dest, target); err != nil {
			return err
		}
	}
}

// Validates and joins an archive path with the destination directory.
//
// Strips leading "./" prefixes common in tar archives, then uses
// [filepath.Localize] to convert slash-separated paths to OS format and
// [filepath.IsLocal] to ensure the path is local (not absolute, no ".."
// traversal, no reserved names on Windows). Returns the validated path
// joined with dest, or ("", nil) for root entries like "./" that resolve
// to the destination itself.
func validateAndJoinPath(dest, name string) (string, error) {
	name = strings.TrimPrefix(name, "./")
	name = strings.TrimRight(name, "/")

	if name == "." || name == "" {
		return "", nil
	}

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
// Handles directories, regular files, symlinks, and hard links. PAX extended
// headers are skipped. Returns [ErrUnsupportedFileType] for all other entry
// types (devices, FIFOs, etc.).
func extractEntry(header *tar.Header, tr *tar.Reader, dest, target string) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, os.FileMode(header.Mode)&os.ModePerm)

	case tar.TypeReg:
		return extractFile(tr, target, os.FileMode(header.Mode)&os.ModePerm)

	case tar.TypeSymlink:
		if err := validateSymlink(dest, target, header.Linkname); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
			return err
		}
		return os.Symlink(header.Linkname, target)

	case tar.TypeLink:
		linkTarget, err := validateHardlink(dest, header.Linkname)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
			return err
		}
		return os.Link(linkTarget, target)

	case tar.TypeXHeader, tar.TypeXGlobalHeader, tar.TypeGNULongName, tar.TypeGNULongLink:
		return nil

	default:
		return ErrUnsupportedFileType
	}
}

// Rejects symlinks whose resolved target escapes the destination tree.
func validateSymlink(dest, target, linkname string) error {
	resolved := linkname
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(filepath.Dir(target), resolved)
	}
	resolved = filepath.Clean(resolved)
	cleanDest := filepath.Clean(dest)
	if resolved != cleanDest && !strings.HasPrefix(resolved, cleanDest+string(filepath.Separator)) {
		return ErrInvalidPath
	}
	return nil
}

// Rejects hard links whose target would escape the destination tree.
//
// Hard link Linkname in a tar archive is an archive-relative path to a
// previously extracted entry. It is validated and joined with dest the
// same way regular entry names are, ensuring the resolved target stays
// within the destination tree.
func validateHardlink(dest, linkname string) (string, error) {
	target, err := validateAndJoinPath(dest, linkname)
	if err != nil {
		return "", err
	}
	if target == "" {
		return "", ErrInvalidPath
	}
	return target, nil
}

// Extracts a regular file from r to target with the given mode.
//
// Creates parent directories as needed with [os.ModePerm], subject to the
// process umask.
func extractFile(r io.Reader, target string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = io.Copy(f, r); err != nil {
		return err
	}

	return nil
}
