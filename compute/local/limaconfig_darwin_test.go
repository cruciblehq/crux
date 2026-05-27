//go:build darwin

package local

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfigTemplate_IsValid(t *testing.T) {
	data := limaConfig{
		Arch:        "aarch64",
		CPUs:        2,
		Memory:      "2GiB",
		Disk:        "10GiB",
		User:        "testuser",
		ImagePath:   "/tmp/vm/machine.qcow2",
		GuestSocket: "/run/containerd/containerd.sock",
		HostSocket:  "/Users/testuser/Library/Caches/crux/instances/local/containerd.sock",
	}

	var buf bytes.Buffer
	if err := limaConfigTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template execution: %v", err)
	}

	output := buf.String()
	required := []string{
		"vmType: vz",
		"arch: aarch64",
		"cpus: 2",
		"memory: 2GiB",
		"disk: 10GiB",
		"mountType: virtiofs",
		"containerd:",
		"system: false",
		"user: false",
		"portForwards:",
		`guestSocket: "/run/containerd/containerd.sock"`,
		`hostSocket: "/Users/testuser/Library/Caches/crux/instances/local/containerd.sock"`,
	}
	for _, s := range required {
		if !strings.Contains(output, s) {
			t.Errorf("config missing %q\nfull output:\n%s", s, output)
		}
	}
}
