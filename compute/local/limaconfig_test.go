//go:build darwin || linux

package local

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"strings"
	"testing"

	speccompute "github.com/cruciblehq/spec/compute"
)

func TestConfigTemplate_ImagePath(t *testing.T) {
	data := limaConfig{
		Arch:      "aarch64",
		CPUs:      2,
		Memory:    "2GiB",
		Disk:      "10GiB",
		User:      "testuser",
		ImagePath: "/tmp/vm/machine.qcow2",
	}

	var buf bytes.Buffer
	if err := limaConfigTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template execution: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "/tmp/vm/machine.qcow2") {
		t.Errorf("expected image path in output:\n%s", output)
	}
}

func TestConfigTemplate_x86(t *testing.T) {
	data := limaConfig{
		Arch:      "x86_64",
		CPUs:      4,
		Memory:    "4GiB",
		Disk:      "20GiB",
		User:      "testuser",
		ImagePath: "/tmp/vm/machine.qcow2",
	}

	var buf bytes.Buffer
	if err := limaConfigTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template execution: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "arch: x86_64") {
		t.Errorf("expected arch x86_64 in output:\n%s", output)
	}
}

func TestBuildLimaConfig_Defaults(t *testing.T) {
	cfg, err := buildLimaConfig("/tmp/test.qcow2")
	if err != nil {
		t.Fatalf("buildLimaConfig: %v", err)
	}

	u, _ := user.Current()

	if cfg.Arch != limaArch() {
		t.Errorf("Arch: got %q, want %q", cfg.Arch, limaArch())
	}
	if cfg.CPUs != defaultLimaCPUs {
		t.Errorf("CPUs: got %d, want %d", cfg.CPUs, defaultLimaCPUs)
	}
	if want := fmt.Sprintf("%dGiB", defaultLimaMemoryGiB); cfg.Memory != want {
		t.Errorf("Memory: got %q, want %q", cfg.Memory, want)
	}
	if want := fmt.Sprintf("%dGiB", defaultLimaDiskGiB); cfg.Disk != want {
		t.Errorf("Disk: got %q, want %q", cfg.Disk, want)
	}
	if cfg.User != u.Username {
		t.Errorf("User: got %q, want %q", cfg.User, u.Username)
	}
	if cfg.UserUID != os.Getuid() {
		t.Errorf("UserUID: got %d, want %d", cfg.UserUID, os.Getuid())
	}
	if cfg.ImagePath != "/tmp/test.qcow2" {
		t.Errorf("ImagePath: got %q, want %q", cfg.ImagePath, "/tmp/test.qcow2")
	}
	if cfg.MachineDir != "/tmp" {
		t.Errorf("MachineDir: got %q, want %q", cfg.MachineDir, "/tmp")
	}
	if cfg.GuestMachineDir != guestMachineDir {
		t.Errorf("GuestMachineDir: got %q, want %q", cfg.GuestMachineDir, guestMachineDir)
	}
	if want := fmt.Sprintf("%s/%s%s", guestMachineDir, cruxdBinaryPrefix, limaArch()); cfg.CruxdBinary != want {
		t.Errorf("CruxdBinary: got %q, want %q", cfg.CruxdBinary, want)
	}
	if cfg.CruxdPort != speccompute.DefaultPort {
		t.Errorf("CruxdPort: got %d, want %d", cfg.CruxdPort, speccompute.DefaultPort)
	}
	if cfg.CruxdHealthPath != speccompute.EndpointHealth {
		t.Errorf("CruxdHealthPath: got %q, want %q", cfg.CruxdHealthPath, speccompute.EndpointHealth)
	}
}
