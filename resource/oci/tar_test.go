package oci

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	godigest "github.com/opencontainers/go-digest"
	ocispecv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// Reads an uncompressed tar stream and returns its entries keyed by name.
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
	if err := WriteScratchTar(&buf); err != nil {
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

	mfstDigest := index.Manifests[0].Digest
	mfstBytes, ok := entries["blobs/sha256/"+mfstDigest.Hex()]
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
	if _, ok := entries["blobs/sha256/"+mfst.Config.Digest.Hex()]; !ok {
		t.Fatalf("config blob %s not found", mfst.Config.Digest)
	}
}

func TestWriteCopyTar(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "top.txt"), []byte("top"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "inner.txt"), []byte("inner"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := WriteCopyTar(&buf, src, "dest"); err != nil {
		t.Fatalf("WriteCopyTar: %v", err)
	}

	entries := readTarEntries(t, buf.Bytes())
	if got := entries["dest/top.txt"]; string(got) != "top" {
		t.Fatalf("dest/top.txt = %q, want %q", got, "top")
	}
	if got := entries["dest/sub/inner.txt"]; string(got) != "inner" {
		t.Fatalf("dest/sub/inner.txt = %q, want %q", got, "inner")
	}
}

func TestWriteCopyTarSingleFile(t *testing.T) {
	src := t.TempDir()
	file := filepath.Join(src, "artifact.bin")
	if err := os.WriteFile(file, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := WriteCopyTar(&buf, file, "dest"); err != nil {
		t.Fatalf("WriteCopyTar: %v", err)
	}

	entries := readTarEntries(t, buf.Bytes())
	if got := entries["dest/artifact.bin"]; string(got) != "data" {
		t.Fatalf("dest/artifact.bin = %q, want %q", got, "data")
	}
}

func TestRewriteTarPaths(t *testing.T) {
	var src bytes.Buffer
	tw := tar.NewWriter(&src)
	for _, name := range []string{"old/file.txt", "old/sub/inner.txt", "other/keep.txt"} {
		if err := tw.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeReg, Size: int64(len(name))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	var dst bytes.Buffer
	if err := RewriteTarPaths(&dst, &src, "/old", "new"); err != nil {
		t.Fatalf("RewriteTarPaths: %v", err)
	}

	got := readTarEntries(t, dst.Bytes())
	want := []string{"new/file.txt", "new/sub/inner.txt", "other/keep.txt"}
	names := make([]string, 0, len(got))
	for n := range got {
		names = append(names, n)
	}
	sort.Strings(names)
	sort.Strings(want)
	if len(names) != len(want) {
		t.Fatalf("got %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("got %v, want %v", names, want)
		}
	}
}

func TestRewriteTarPathsSingleFile(t *testing.T) {
	var src bytes.Buffer
	tw := tar.NewWriter(&src)
	if err := tw.WriteHeader(&tar.Header{Name: "./build.img", Typeflag: tar.TypeReg, Size: 2}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("ok")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	var dst bytes.Buffer
	if err := RewriteTarPaths(&dst, &src, "build.img", "disk/root.img"); err != nil {
		t.Fatalf("RewriteTarPaths: %v", err)
	}

	got := readTarEntries(t, dst.Bytes())
	if _, ok := got["disk/root.img"]; !ok {
		t.Fatalf("renamed entry missing, got %v", got)
	}
}
