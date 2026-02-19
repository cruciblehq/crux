package archive

import (
	"archive/tar"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
)

// Creates a compressed tar archive from a directory.
//
// The compression format is detected from the dest filename extension (see
// [Format]). The archive contains all files and directories under src with
// paths stored relative to src. Paths in the archive use forward slashes
// regardless of the host operating system. Only regular files and directories
// are archivable; symlinks and other special file types such as devices and
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

	// Override name and strip special bits (setuid, setgid, sticky)
	header.Name = filepath.ToSlash(relPath)
	header.Mode = int64(info.Mode().Perm())

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
