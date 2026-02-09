package oci

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/paths"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
)

const (

	// Keep the original file modification time (see [writeTarEntry]).
	preserveModTime = false

	// Set modification time to Unix epoch for reproducible builds (see [writeTarEntry]).
	zeroModTime = true
)

// Writes an image index to a tarball by creating a temporary OCI layout.
//
// The go-containerregistry tarball package only supports single images, so this
// function writes to an OCI layout directory first, then archives it.
func writeIndexToTarball(idx v1.ImageIndex, path string) error {
	tmpDir, err := os.MkdirTemp("", "oci-layout-*")
	if err != nil {
		return crex.Wrap(ErrLayoutWrite, err)
	}
	defer os.RemoveAll(tmpDir)

	layoutPath := filepath.Join(tmpDir, "oci")
	if _, err := layout.Write(layoutPath, idx); err != nil {
		return crex.Wrap(ErrLayoutWrite, err)
	}

	return writeDirToTarball(layoutPath, path)
}

// Writes a directory to a tarball file.
func writeDirToTarball(srcDir, destPath string) error {
	f, err := os.Create(destPath)
	if err != nil {
		return crex.Wrap(ErrLayoutWrite, err)
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
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

		return writeTarEntry(tw, path, relPath, info, preserveModTime)
	})
}

// Writes a single entry to a tar writer.
//
// When zeroModTime is true, the modification time is set to the Unix epoch
// for reproducible builds. This is used for layer creation where consistent
// timestamps are required.
func writeTarEntry(tw *tar.Writer, srcPath, name string, info os.FileInfo, zeroModTime bool) error {
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = name
	if zeroModTime {
		header.ModTime = time.Time{}
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(tw, file)
	return err
}

// Creates a tar archive from directory contents.
func createTarFromDir(srcDir, destDir string) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)
		destPath = filepath.ToSlash(destPath)

		if destPath == destDir || destPath == "." {
			return nil
		}

		return writeTarEntry(tw, path, destPath, info, zeroModTime)
	})

	if err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Creates a tar archive containing a single file.
func createTarFromBytes(content []byte, dest string, mode int64) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	header := &tar.Header{
		Name:    filepath.ToSlash(dest),
		Size:    int64(len(content)),
		Mode:    mode,
		ModTime: time.Time{},
	}

	if err := tw.WriteHeader(header); err != nil {
		return nil, err
	}
	if _, err := tw.Write(content); err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Extracts a tar archive to a destination directory.
//
// Validates paths to prevent directory traversal attacks.
func extractTar(r io.Reader, destDir string) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return crex.Wrap(ErrInvalidImage, err)
		}

		if err := extractTarEntry(tr, header, destDir); err != nil {
			return err
		}
	}
}

// Extracts a single tar entry to the destination directory.
func extractTarEntry(r io.Reader, header *tar.Header, destDir string) error {
	localName, err := filepath.Localize(header.Name)
	if err != nil || !filepath.IsLocal(localName) {
		return crex.Wrap(ErrInvalidImage, ErrInvalidTarPath)
	}
	target := filepath.Join(destDir, localName)

	switch header.Typeflag {
	case tar.TypeDir:
		if err := os.MkdirAll(target, paths.DefaultDirMode); err != nil {
			return crex.Wrap(ErrInvalidImage, err)
		}
	case tar.TypeReg:
		if err := os.MkdirAll(filepath.Dir(target), paths.DefaultDirMode); err != nil {
			return crex.Wrap(ErrInvalidImage, err)
		}
		if err := writeFile(target, r); err != nil {
			return err
		}
	}
	return nil
}

// Writes content from a reader to a file.
func writeFile(path string, r io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return crex.Wrap(ErrInvalidImage, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return crex.Wrap(ErrInvalidImage, err)
	}
	return nil
}
