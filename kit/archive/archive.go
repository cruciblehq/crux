package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cruciblehq/crux/kit/crex"
	"github.com/klauspost/compress/zstd"
)

// Supported archive compression formats.
//
// Each format corresponds to a tar archive compressed with a specific
// algorithm. The format is inferred from the file extension by [Create] and
// [Extract], or supplied explicitly to [ExtractFromReader].
type Format int

const (

	// Zstandard compression (.tar.zst).
	Zstd Format = iota

	// Gzip compression (.tar.gz, .tgz).
	Gzip

	// File extension for Zstandard-compressed tar archives.
	extZstd = ".tar.zst"

	// File extension for Gzip-compressed tar archives.
	extGzip = ".tar.gz"

	// Alternate file extension for Gzip-compressed tar archives.
	extTgz = ".tgz"
)

// String returns the canonical file extension for the format.
func (f Format) String() string {
	switch f {
	case Zstd:
		return extZstd
	case Gzip:
		return extGzip
	default:
		return extZstd
	}
}

// Detects the archive format from a filename.
//
// Returns [ErrUnsupportedFormat] if the extension is not recognised.
func detect(name string) (Format, error) {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, extZstd):
		return Zstd, nil
	case strings.HasSuffix(lower, extGzip):
		return Gzip, nil
	case strings.HasSuffix(lower, extTgz):
		return Gzip, nil
	default:
		return 0, ErrUnsupportedFormat
	}
}

// Returns a write-closer that compresses data with the given format.
func newCompressWriter(w io.Writer, f Format) (io.WriteCloser, error) {
	switch f {
	case Zstd:
		return zstd.NewWriter(w)
	case Gzip:
		return gzip.NewWriter(w), nil
	default:
		return nil, ErrUnsupportedFormat
	}
}

// Returns a read-closer that decompresses data with the given format.
func newDecompressReader(r io.Reader, f Format) (io.ReadCloser, error) {
	switch f {
	case Zstd:
		zr, err := zstd.NewReader(r)
		if err != nil {
			return nil, err
		}
		return zr.IOReadCloser(), nil
	case Gzip:
		return gzip.NewReader(r)
	default:
		return nil, ErrUnsupportedFormat
	}
}

const (

	// Permission mode used when creating directories.
	//
	// This mode is required when handling resource extraction and storage and
	// optional for other purposes.
	DirMode os.FileMode = 0755

	// Permission mode used when creating files during archive creation.
	//
	// During extraction, file modes are preserved from the archive headers
	// with special bits (setuid, setgid, sticky) stripped.
	FileMode os.FileMode = 0644
)

// Creates a compressed tar archive from a directory.
//
// The compression format is detected from the dest filename extension
// (see [Format]). The archive contains all files and directories under src
// with paths stored relative to src. Paths in the archive use forward slashes
// regardless of the host operating system. Only regular files and directories
// are stored. Symlinks and other special file types such as devices and
// sockets will cause the function to return [ErrUnsupportedFileType]. If
// creation fails, the partially written archive is removed.
func Create(src, dest string) (err error) {
	fmt, err := detect(dest)
	if err != nil {
		return crex.Wrap(ErrCreateFailed, err)
	}

	file, err := os.Create(dest)
	if err != nil {
		return crex.Wrap(ErrCreateFailed, err)
	}
	defer file.Close()

	cw, err := newCompressWriter(file, fmt)
	if err != nil {
		os.Remove(dest)
		return crex.Wrap(ErrCreateFailed, err)
	}
	defer func() {
		cw.Close()
		if err != nil {
			os.Remove(dest)
		}
	}()

	tw := tar.NewWriter(cw)
	defer tw.Close()

	if err = writeTar(tw, src); err != nil {
		return crex.Wrap(ErrCreateFailed, err)
	}

	return nil
}

// Extracts a compressed tar archive to a directory.
//
// The compression format is detected from the src filename extension (see
// [Format]). File permissions are preserved from the archive headers with
// special bits stripped. Directories are created with [DirMode]. Returns
// [ErrExtractFailed] wrapping [os.ErrExist] if dest already exists. Regular
// files, directories, symlinks, and hard links are supported. Symlinks and
// hard links are validated to ensure they do not escape the destination tree.
// PAX extended headers are skipped transparently. Other entry types such as
// devices and sockets return [ErrUnsupportedFileType]. Absolute paths and
// path traversal attempts (e.g., "../etc/passwd") return [ErrInvalidPath].
// If extraction fails, the destination directory and its contents are removed.
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
// to detect from. File permissions are preserved from the archive headers
// with special bits stripped. Supports the same entry types as [Extract]:
// regular files, directories, symlinks, and hard links (all validated
// against directory escape). PAX headers are skipped. Unlike [Extract],
// this function does not check whether dest already exists and does not
// clean up on failure.
func ExtractFromReader(r io.Reader, dest string, f Format) error {
	dr, err := newDecompressReader(r, f)
	if err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}
	defer dr.Close()

	if err := os.MkdirAll(dest, DirMode); err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}

	if err := readTar(tar.NewReader(dr), dest); err != nil {
		return crex.Wrap(ErrExtractFailed, err)
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
	header.Mode = int64(FileMode)
	if info.IsDir() {
		header.Mode = int64(DirMode)
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
		return os.MkdirAll(target, DirMode)

	case tar.TypeReg:
		return extractFile(tr, target, os.FileMode(header.Mode)&0777)

	case tar.TypeSymlink:
		if err := validateSymlink(dest, target, header.Linkname); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), DirMode); err != nil {
			return err
		}
		return os.Symlink(header.Linkname, target)

	case tar.TypeLink:
		linkTarget, err := validateHardlink(dest, header.Linkname)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), DirMode); err != nil {
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
// Creates parent directories as needed.
func extractFile(r io.Reader, target string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), DirMode); err != nil {
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

// Reads a named file from a tar archive.
//
// Scans the tar reader sequentially until filename is found or the archive is
// exhausted. Returns the file contents and nil error on success, (nil, nil) if
// the file is not present, or (nil, error) if a read error occurs. The tar
// reader is advanced past the matched entry and cannot be rewound.
func Find(tr *tar.Reader, filename string) ([]byte, error) {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}

		if header.Name == filename {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
	}
}
