//go:build darwin

package vm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractLimactl(t *testing.T) {
	// Create a fake tar.gz archive with a bin/limactl entry.
	var archiveBuf bytes.Buffer
	gw := gzip.NewWriter(&archiveBuf)
	tw := tar.NewWriter(gw)

	content := []byte("#!/bin/sh\necho fake limactl")
	tw.WriteHeader(&tar.Header{
		Name:     "bin/limactl",
		Mode:     0755,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	})
	tw.Write(content)
	tw.Close()
	gw.Close()

	// Extract to temp dir.
	tmpDir := t.TempDir()
	dest := filepath.Join(tmpDir, "bin", "limactl")
	if err := extractLimactl(&archiveBuf, dest); err != nil {
		t.Fatalf("extractLimactl: %v", err)
	}

	// Verify the file was extracted.
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading extracted file: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("expected %q, got %q", string(content), string(got))
	}

	// Verify it's executable.
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("expected executable permissions")
	}
}
