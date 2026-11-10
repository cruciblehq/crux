package oci

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
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
	if _, ok := entries[fmt.Sprintf("blobs/sha256/%s", mfst.Config.Digest.Hex())]; !ok {
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

// Reads a tar stream and returns its headers keyed by name.
func readTarHeaders(t *testing.T, data []byte) map[string]*tar.Header {
	t.Helper()
	headers := make(map[string]*tar.Header)
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("read tar: %v", err)
		}
		headers[hdr.Name] = hdr
	}
	return headers
}

// Writes a gzip-compressed tar of the given headers to a temp file and returns its path.
func writeGzTarFile(t *testing.T, headers []*tar.Header, bodies map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rootfs.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for _, h := range headers {
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if b, ok := bodies[h.Name]; ok {
			if _, err := tw.Write([]byte(b)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestWriteUnpackTarRootGzip(t *testing.T) {
	path := writeGzTarFile(t, []*tar.Header{
		{Name: "bin/sh", Typeflag: tar.TypeReg, Mode: 0o755, Size: 2},
		{Name: "bin/bash", Typeflag: tar.TypeSymlink, Linkname: "sh", Mode: 0o777},
		{Name: "./etc/hostname", Typeflag: tar.TypeReg, Mode: 0o644, Size: 2},
	}, map[string]string{"bin/sh": "#!", "./etc/hostname": "hi"})

	var buf bytes.Buffer
	if err := WriteUnpackTar(&buf, path, "/"); err != nil {
		t.Fatalf("WriteUnpackTar: %v", err)
	}

	headers := readTarHeaders(t, buf.Bytes())

	// At the root, entries are emitted unprefixed and "./" is stripped.
	if _, ok := headers["bin/sh"]; !ok {
		t.Fatalf("bin/sh missing, got %v keys", keysOf(headers))
	}
	if _, ok := headers["etc/hostname"]; !ok {
		t.Fatalf("etc/hostname missing, got %v keys", keysOf(headers))
	}
	// Symlink header and its target are preserved.
	sym, ok := headers["bin/bash"]
	if !ok {
		t.Fatalf("bin/bash symlink missing")
	}
	if sym.Typeflag != tar.TypeSymlink || sym.Linkname != "sh" {
		t.Fatalf("symlink not preserved: type=%d link=%q", sym.Typeflag, sym.Linkname)
	}
}

func TestWriteUnpackTarDestPrefix(t *testing.T) {
	path := writeGzTarFile(t, []*tar.Header{
		{Name: "lib/data.txt", Typeflag: tar.TypeReg, Mode: 0o644, Size: 2},
	}, map[string]string{"lib/data.txt": "ok"})

	var buf bytes.Buffer
	if err := WriteUnpackTar(&buf, path, "/opt/app"); err != nil {
		t.Fatalf("WriteUnpackTar: %v", err)
	}

	entries := readTarEntries(t, buf.Bytes())
	if got := entries["opt/app/lib/data.txt"]; string(got) != "ok" {
		t.Fatalf("opt/app/lib/data.txt = %q, want %q", got, "ok")
	}
}

func TestWriteUnpackTarUncompressed(t *testing.T) {
	// A plain (non-gzip) tar must also be accepted.
	path := filepath.Join(t.TempDir(), "rootfs.tar")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(f)
	if err := tw.WriteHeader(&tar.Header{Name: "usr/bin/tool", Typeflag: tar.TypeReg, Mode: 0o755, Size: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	f.Close()

	var buf bytes.Buffer
	if err := WriteUnpackTar(&buf, path, "/"); err != nil {
		t.Fatalf("WriteUnpackTar: %v", err)
	}
	entries := readTarEntries(t, buf.Bytes())
	if got := entries["usr/bin/tool"]; string(got) != "x" {
		t.Fatalf("usr/bin/tool = %q, want %q", got, "x")
	}
}

// Returns the sorted keys of a header map for diagnostics.
func keysOf(m map[string]*tar.Header) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
