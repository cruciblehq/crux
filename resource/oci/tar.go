package oci

import (
	"archive/tar"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	godigest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go"
	ocispecv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/cruciblehq/utils-go/crex"
)

// Writes a minimal OCI image tar representing a scratch (empty) image.
//
// The archive contains an OCI layout marker, a minimal config blob with no
// layers, a manifest referencing the config, and an index referencing the
// manifest.
func WriteScratchTar(w io.Writer) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	layout, err := json.Marshal(ocispecv1.ImageLayout{Version: ocispecv1.ImageLayoutVersion})
	if err != nil {
		return err
	}
	if err := writeTarEntry(tw, ocispecv1.ImageLayoutFile, layout); err != nil {
		return err
	}

	configDigest, configSize, err := writeScratchConfig(tw)
	if err != nil {
		return err
	}

	mfstDigest, mfstSize, err := writeScratchManifest(tw, configDigest, configSize)
	if err != nil {
		return err
	}

	return writeScratchIndex(tw, mfstDigest, mfstSize)
}

// Writes the scratch image config blob and returns its digest and byte length.
//
// Kept as a raw JSON literal to avoid encoding/json emitting "config":{} from
// the zero-value ImageConfig field in ocispecv1.Image.
func writeScratchConfig(tw *tar.Writer) (godigest.Digest, int64, error) {
	config := []byte(`{"architecture":"amd64","os":"linux","rootfs":{"type":"layers","diff_ids":[]}}`)
	d, err := writeBlob(tw, config)
	return d, int64(len(config)), err
}

// Writes the image manifest blob referencing config and returns its digest and byte length.
func writeScratchManifest(tw *tar.Writer, configDigest godigest.Digest, configSize int64) (godigest.Digest, int64, error) {
	mfst, err := json.Marshal(ocispecv1.Manifest{
		Versioned: ocispec.Versioned{SchemaVersion: 2},
		MediaType: ocispecv1.MediaTypeImageManifest,
		Config: ocispecv1.Descriptor{
			MediaType: ocispecv1.MediaTypeImageConfig,
			Digest:    configDigest,
			Size:      configSize,
		},
		Layers: []ocispecv1.Descriptor{},
	})
	if err != nil {
		return "", 0, err
	}
	d, err := writeBlob(tw, mfst)
	return d, int64(len(mfst)), err
}

// Writes the image index referencing the single manifest.
func writeScratchIndex(tw *tar.Writer, mfstDigest godigest.Digest, mfstSize int64) error {
	index, err := json.Marshal(ocispecv1.Index{
		Versioned: ocispec.Versioned{SchemaVersion: 2},
		MediaType: ocispecv1.MediaTypeImageIndex,
		Manifests: []ocispecv1.Descriptor{
			{
				MediaType: ocispecv1.MediaTypeImageManifest,
				Digest:    mfstDigest,
				Size:      mfstSize,
			},
		},
	})
	if err != nil {
		return err
	}
	return writeTarEntry(tw, ocispecv1.ImageIndexFile, index)
}

// Writes data as a content-addressed blob under blobs/sha256/<digest> and returns the digest.
func writeBlob(tw *tar.Writer, data []byte) (godigest.Digest, error) {
	d := godigest.FromBytes(data)
	return d, writeTarEntry(tw, ocispecv1.ImageBlobsDir+"/sha256/"+d.Hex(), data)
}

// Writes an in-memory byte slice as a regular tar entry.
func writeTarEntry(tw *tar.Writer, name string, data []byte) error {
	hdr := &tar.Header{
		Name:     name,
		Mode:     0644,
		Size:     int64(len(data)),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

// Writes host files at srcPath into a tar archive destined for destDir.
//
// Directories are walked recursively; symlinks are followed by Lstat and
// written as regular paths. The destination paths inside the tar are formed by
// joining destDir with the relative path from srcPath.
func WriteCopyTar(w io.Writer, srcPath, destDir string) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	info, err := os.Stat(srcPath)
	if err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}

	if !info.IsDir() {
		tarPath := filepath.Join(destDir, filepath.Base(srcPath))
		return copyFileToTar(tw, srcPath, tarPath)
	}

	return filepath.WalkDir(srcPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return crex.Wrap(ErrFileSystemOperation, err)
		}
		rel, err := filepath.Rel(srcPath, path)
		if err != nil {
			return crex.Wrap(ErrFileSystemOperation, err)
		}
		tarPath := filepath.Join(destDir, rel)
		if d.IsDir() {
			return tw.WriteHeader(&tar.Header{
				Name:     tarPath + "/",
				Typeflag: tar.TypeDir,
				Mode:     0755,
			})
		}
		return copyFileToTar(tw, path, tarPath)
	})
}

// Copies a single file from the host into a tar writer at tarPath.
func copyFileToTar(tw *tar.Writer, hostPath, tarPath string) error {
	fi, err := os.Lstat(hostPath)
	if err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	hdr, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	hdr.Name = tarPath

	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	f, err := os.Open(hostPath)
	if err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	defer f.Close()

	_, err = io.Copy(tw, f)
	return err
}

// Rewrites entry paths in an uncompressed tar stream from srcPath to destPath.
//
// When srcPath names a single file the entry is renamed to destPath. When
// srcPath names a directory, the srcPath prefix in every entry name is replaced
// by destPath. Both paths must use forward slashes. Leading slashes in srcPath
// and destPath are stripped before comparison and output.
func RewriteTarPaths(w io.Writer, r io.Reader, srcPath, destPath string) error {
	srcClean := strings.Trim(srcPath, "/")
	destClean := strings.Trim(destPath, "/")

	tr := tar.NewReader(r)
	tw := tar.NewWriter(w)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		name := strings.TrimPrefix(strings.TrimPrefix(hdr.Name, "./"), "/")

		var newName string
		switch {
		case name == srcClean:
			newName = destClean
		case strings.HasPrefix(name, srcClean+"/"):
			newName = destClean + "/" + strings.TrimPrefix(name, srcClean+"/")
		default:
			newName = name
		}

		hdr.Name = newName
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := io.Copy(tw, tr); err != nil {
			return err
		}
	}
	return tw.Close()
}
