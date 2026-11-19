package oci

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	godigest "github.com/opencontainers/go-digest"
	ocispecv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func readTarEntries(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
	entries := make(map[string][]byte)
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("read tar: %v", err)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("read entry %q: %v", hdr.Name, err)
		}
		entries[hdr.Name] = body
	}
	return entries
}

func TestWriteScratchTar(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteScratchTar(&buf, "arm64"); err != nil {
		t.Fatalf("WriteScratchTar: %v", err)
	}
	entries := readTarEntries(t, buf.Bytes())

	if _, ok := entries[ocispecv1.ImageLayoutFile]; !ok {
		t.Fatalf("missing %s", ocispecv1.ImageLayoutFile)
	}

	indexBytes, ok := entries[ocispecv1.ImageIndexFile]
	if !ok {
		t.Fatalf("missing %s", ocispecv1.ImageIndexFile)
	}
	var index ocispecv1.Index
	if err := json.Unmarshal(indexBytes, &index); err != nil {
		t.Fatalf("decode index: %v", err)
	}
	if len(index.Manifests) != 1 {
		t.Fatalf("want 1 manifest, got %d", len(index.Manifests))
	}

	if got := index.Manifests[0].Annotations[annotationImageName]; got != scratchImageRef {
		t.Fatalf("manifest %s annotation = %q, want %q", annotationImageName, got, scratchImageRef)
	}

	mfstDigest := index.Manifests[0].Digest
	mfstBytes, ok := entries[fmt.Sprintf("blobs/sha256/%s", mfstDigest.Hex())]
	if !ok {
		t.Fatalf("manifest blob %s not found", mfstDigest)
	}
	if godigest.FromBytes(mfstBytes) != mfstDigest {
		t.Fatalf("manifest digest mismatch")
	}

	var mfst ocispecv1.Manifest
	if err := json.Unmarshal(mfstBytes, &mfst); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if len(mfst.Layers) != 0 {
		t.Fatalf("scratch image must have no layers, got %d", len(mfst.Layers))
	}
	configBytes, ok := entries[fmt.Sprintf("blobs/sha256/%s", mfst.Config.Digest.Hex())]
	if !ok {
		t.Fatalf("config blob %s not found", mfst.Config.Digest)
	}
	var cfg ocispecv1.Image
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		t.Fatalf("decode config: %v", err)
	}
	if cfg.Architecture != "arm64" || cfg.OS != "linux" {
		t.Fatalf("platform = %s/%s, want linux/arm64", cfg.OS, cfg.Architecture)
	}
}
