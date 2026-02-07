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

func TestExtractLima(t *testing.T) {
	// Create a fake tar.gz archive with a bin/limactl entry and a guest agent.
	var archiveBuf bytes.Buffer
	gw := gzip.NewWriter(&archiveBuf)
	tw := tar.NewWriter(gw)

	limactlContent := []byte("#!/bin/sh\necho fake limactl")
	tw.WriteHeader(&tar.Header{
		Name:     "bin/limactl",
		Mode:     0755,
		Size:     int64(len(limactlContent)),
		Typeflag: tar.TypeReg,
	})
	tw.Write(limactlContent)

	agentContent := []byte("fake-guest-agent")
	tw.WriteHeader(&tar.Header{
		Name:     "share/lima/lima-guestagent.Linux-aarch64.gz",
		Mode:     0644,
		Size:     int64(len(agentContent)),
		Typeflag: tar.TypeReg,
	})
	tw.Write(agentContent)

	tw.Close()
	gw.Close()

	// Extract to temp dir.
	tmpDir := t.TempDir()
	if err := extractLima(&archiveBuf, tmpDir); err != nil {
		t.Fatalf("extractLima: %v", err)
	}

	// Verify limactl was extracted.
	limactlPath := filepath.Join(tmpDir, "bin", "limactl")
	got, err := os.ReadFile(limactlPath)
	if err != nil {
		t.Fatalf("reading extracted limactl: %v", err)
	}
	if string(got) != string(limactlContent) {
		t.Errorf("limactl: expected %q, got %q", string(limactlContent), string(got))
	}

	// Verify it's executable.
	info, err := os.Stat(limactlPath)
	if err != nil {
		t.Fatalf("stat limactl: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("expected executable permissions on limactl")
	}

	// Verify guest agent was also extracted.
	agentPath := filepath.Join(tmpDir, "share", "lima", "lima-guestagent.Linux-aarch64.gz")
	gotAgent, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("reading extracted guest agent: %v", err)
	}
	if string(gotAgent) != string(agentContent) {
		t.Errorf("guest agent: expected %q, got %q", string(agentContent), string(gotAgent))
	}
}
