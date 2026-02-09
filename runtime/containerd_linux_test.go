//go:build linux

package runtime

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractContainerd(t *testing.T) {
	var archiveBuf bytes.Buffer
	gw := gzip.NewWriter(&archiveBuf)
	tw := tar.NewWriter(gw)

	containerdContent := []byte("#!/bin/sh\necho fake containerd")
	tw.WriteHeader(&tar.Header{
		Name:     "bin/containerd",
		Mode:     0755,
		Size:     int64(len(containerdContent)),
		Typeflag: tar.TypeReg,
	})
	tw.Write(containerdContent)

	shimContent := []byte("fake-shim")
	tw.WriteHeader(&tar.Header{
		Name:     "bin/containerd-shim-runc-v2",
		Mode:     0755,
		Size:     int64(len(shimContent)),
		Typeflag: tar.TypeReg,
	})
	tw.Write(shimContent)

	tw.Close()
	gw.Close()

	tmpDir := t.TempDir()
	if err := extractContainerd(&archiveBuf, tmpDir); err != nil {
		t.Fatalf("extractContainerd: %v", err)
	}

	binPath := filepath.Join(tmpDir, "bin", "containerd")
	got, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("reading extracted containerd: %v", err)
	}
	if string(got) != string(containerdContent) {
		t.Errorf("containerd: expected %q, got %q", string(containerdContent), string(got))
	}

	info, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("stat containerd: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("expected executable permissions on containerd")
	}

	shimPath := filepath.Join(tmpDir, "bin", "containerd-shim-runc-v2")
	gotShim, err := os.ReadFile(shimPath)
	if err != nil {
		t.Fatalf("reading extracted shim: %v", err)
	}
	if string(gotShim) != string(shimContent) {
		t.Errorf("shim: expected %q, got %q", string(shimContent), string(gotShim))
	}
}

func TestExtractContainerd_MissingBinary(t *testing.T) {
	var archiveBuf bytes.Buffer
	gw := gzip.NewWriter(&archiveBuf)
	tw := tar.NewWriter(gw)

	otherContent := []byte("some other file")
	tw.WriteHeader(&tar.Header{
		Name:     "bin/other",
		Mode:     0755,
		Size:     int64(len(otherContent)),
		Typeflag: tar.TypeReg,
	})
	tw.Write(otherContent)

	tw.Close()
	gw.Close()

	tmpDir := t.TempDir()
	err := extractContainerd(&archiveBuf, tmpDir)
	if err == nil {
		t.Fatal("expected error when containerd binary is missing")
	}
}

func TestGenerateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	t.Setenv("XDG_RUNTIME_DIR", filepath.Join(tmpDir, "run"))

	configPath, err := generateConfig()
	if err != nil {
		t.Fatalf("generateConfig: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}

	output := string(content)
	required := []string{
		"version = 3",
		"root =",
		"state =",
		"address =",
		"containerd.sock",
	}
	for _, s := range required {
		if !strings.Contains(output, s) {
			t.Errorf("config missing %q\nfull output:\n%s", s, output)
		}
	}
}
