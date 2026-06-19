//go:build darwin || linux

package local

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"slices"
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

func TestLimaURL(t *testing.T) {
	url := limaURL()

	for _, sub := range []string{limaVersion, limaOS(), limaDownloadArch()} {
		if !strings.Contains(url, sub) {
			t.Errorf("limaURL() missing %q; full URL: %s", sub, url)
		}
	}
	if !strings.HasPrefix(url, "https://") {
		t.Errorf("limaURL() not https: %s", url)
	}
	if !strings.HasSuffix(url, ".tar.gz") {
		t.Errorf("limaURL() missing .tar.gz suffix: %s", url)
	}
}

func TestLimaEnv_ContainsLimaHome(t *testing.T) {
	env := limaEnv()
	want := "LIMA_HOME=" + vmDir()
	if slices.Contains(env, want) {
		return
	}
	t.Errorf("limaEnv missing %q; got %v", want, env)
}

func TestLimaEnv_IncludesSetVars(t *testing.T) {
	t.Setenv("HOME", "/tmp/testhome")
	t.Setenv("USER", "testuser")
	t.Setenv("PATH", "/usr/bin")
	t.Setenv("TMPDIR", "/tmp")

	env := limaEnv()
	for _, want := range []string{
		"HOME=/tmp/testhome",
		"USER=testuser",
		"PATH=/usr/bin",
		"TMPDIR=/tmp",
	} {
		found := false
		for _, e := range env {
			if e == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("limaEnv missing %q; got %v", want, env)
		}
	}
}

func TestLimaEnv_ExcludesEmptyVars(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USER", "")
	t.Setenv("PATH", "")
	t.Setenv("TMPDIR", "")

	env := limaEnv()
	for _, e := range env {
		for _, key := range []string{"HOME", "USER", "PATH", "TMPDIR"} {
			if strings.HasPrefix(e, key+"=") {
				t.Errorf("limaEnv included %q when var was empty", e)
			}
		}
	}
}
