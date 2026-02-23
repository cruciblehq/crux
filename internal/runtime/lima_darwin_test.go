//go:build darwin

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

func TestConfigTemplate_IsValid(t *testing.T) {
	data := configData{
		Arch:          "aarch64",
		CPUs:          2,
		Memory:        "2GiB",
		Disk:          "10GiB",
		GuestSocket:   "/run/cruxd/cruxd.sock",
		HostSocket:    "/tmp/test/cruxd.sock",
		User:          "testuser",
		ContainerdGID: 999,
		CruxdGID:      989,
	}

	var buf bytes.Buffer
	if err := configTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template execution: %v", err)
	}

	output := buf.String()
	required := []string{
		"vmType: vz",
		"arch: aarch64",
		"cpus: 2",
		"memory: 2GiB",
		"disk: 10GiB",
		"containerd:",
		"system: false",
		"user: false",
		"apk add --no-cache containerd",
		"addgroup -g 989 -S cruxd",
		"addgroup testuser cruxd",
		`command_user="testuser"`,
		"chown testuser:cruxd /run/cruxd",
		"/run/cruxd/cruxd.sock",
		"/tmp/test/cruxd.sock",
	}
	for _, s := range required {
		if !strings.Contains(output, s) {
			t.Errorf("config missing %q\nfull output:\n%s", s, output)
		}
	}
}

func TestConfigTemplate_x86(t *testing.T) {
	data := configData{
		Arch:          "x86_64",
		CPUs:          4,
		Memory:        "4GiB",
		Disk:          "20GiB",
		GuestSocket:   "/run/cruxd/cruxd.sock",
		HostSocket:    "/tmp/test/cruxd.sock",
		User:          "testuser",
		ContainerdGID: 999,
		CruxdGID:      989,
	}

	var buf bytes.Buffer
	if err := configTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template execution: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "arch: x86_64") {
		t.Errorf("expected arch x86_64 in output:\n%s", output)
	}
}
