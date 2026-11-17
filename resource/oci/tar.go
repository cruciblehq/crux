package oci

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"

	godigest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go"
	ocispecv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/cruciblehq/utils-go/file"
)

// Names for the scratch base image.
//
// containerd's default Import only registers a manifest if it carries the
// "io.containerd.image.name" annotation. The standard ref-name annotation
// is also set for tools that key on it instead.
const (
	annotationImageName = "io.containerd.image.name" // containerd image-name annotation key.
	scratchImageRef     = "scratch:latest"           // Full name registered for the scratch image.
	scratchImageTag     = "latest"                   // OCI ref-name tag for the scratch image.
)

// Schema version of OCI image manifests and indexes.
const ociSchemaVersion = 2

// Writes scratch (empty) OCI image.
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
		Versioned: ocispec.Versioned{SchemaVersion: ociSchemaVersion},
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
//
// The manifest descriptor carries an image-name annotation so containerd's
// default Import registers a named image. Without it, Import silently discards
// the unnamed manifest and reports no images.
func writeScratchIndex(tw *tar.Writer, mfstDigest godigest.Digest, mfstSize int64) error {
	index, err := json.Marshal(ocispecv1.Index{
		Versioned: ocispec.Versioned{SchemaVersion: ociSchemaVersion},
		MediaType: ocispecv1.MediaTypeImageIndex,
		Manifests: []ocispecv1.Descriptor{
			{
				MediaType: ocispecv1.MediaTypeImageManifest,
				Digest:    mfstDigest,
				Size:      mfstSize,
				Annotations: map[string]string{
					annotationImageName:         scratchImageRef,
					ocispecv1.AnnotationRefName: scratchImageTag,
				},
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
	return d, writeTarEntry(tw, fmt.Sprintf("%s/sha256/%s", ocispecv1.ImageBlobsDir, d.Hex()), data)
}

// Writes an in-memory byte slice as a regular tar entry.
func writeTarEntry(tw *tar.Writer, name string, data []byte) error {
	hdr := &tar.Header{
		Name:     name,
		Mode:     int64(file.DefaultFileMode),
		Size:     int64(len(data)),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}
