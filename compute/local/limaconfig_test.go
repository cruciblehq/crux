//go:build darwin || linux

package local

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"strings"
	"testing"

	"github.com/cruciblehq/crux/files"
	"github.com/cruciblehq/crux/security/vm"
)

func TestConfigTemplate_ImagePath(t *testing.T) {
	data := limaConfig{
		Arch:        "aarch64",
		CPUs:        2,
		Memory:      "2GiB",
		Disk:        "10GiB",
		User:        "testuser",
		ImagePath:   "/tmp/vm/machine.qcow2",
		GuestSocket: "/run/containerd/containerd.sock",
		HostSocket:  "/home/testuser/.cache/crux/instances/local/containerd.sock",
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
		Arch:        "x86_64",
		CPUs:        4,
		Memory:      "4GiB",
		Disk:        "20GiB",
		User:        "testuser",
		ImagePath:   "/tmp/vm/machine.qcow2",
		GuestSocket: "/run/containerd/containerd.sock",
		HostSocket:  "/home/testuser/.cache/crux/instances/local/containerd.sock",
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
	cfg, err := buildLimaConfig("/tmp/test.qcow2", vm.VM{})
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
	if cfg.GuestSocket != guestContainerdSocket {
		t.Errorf("GuestSocket: got %q, want %q", cfg.GuestSocket, guestContainerdSocket)
	}
	if want := files.ContainerdSocket(limaInstanceName); cfg.HostSocket != want {
		t.Errorf("HostSocket: got %q, want %q", cfg.HostSocket, want)
	}
	if len(cfg.Sysctls) != 0 {
		t.Errorf("Sysctls: expected empty, got %v", cfg.Sysctls)
	}
	if len(cfg.NftRules) != 0 {
		t.Errorf("NftRules: expected empty, got %v", cfg.NftRules)
	}
}

func TestBuildLimaConfig_PropagatesPolicy(t *testing.T) {
	vmSpec := vm.VM{
		Sysctls: map[string]string{"net.ipv4.ip_forward": "1"},
		Nftables: []vm.VMNftRule{
			{Table: "inet filter", Chain: "input", Rule: "drop"},
		},
	}

	cfg, err := buildLimaConfig("/tmp/test.qcow2", vmSpec)
	if err != nil {
		t.Fatalf("buildLimaConfig: %v", err)
	}

	if got := cfg.Sysctls["net.ipv4.ip_forward"]; got != "1" {
		t.Errorf("Sysctls[net.ipv4.ip_forward]: got %q, want %q", got, "1")
	}
	if len(cfg.NftRules) != 1 {
		t.Fatalf("NftRules: got %d entries, want 1", len(cfg.NftRules))
	}
	r := cfg.NftRules[0]
	if r.Table != "inet filter" || r.Chain != "input" || r.Rule != "drop" {
		t.Errorf("NftRules[0]: got %+v", r)
	}
}
