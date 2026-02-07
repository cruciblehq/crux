//go:build darwin

package vm

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfigTemplate_IsValid(t *testing.T) {
	data := configData{
		Arch:    "aarch64",
		CPUs:    2,
		Memory:  "2GiB",
		Disk:    "10GiB",
		DataDir: "/tmp/crux-test",
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
		"mountType: virtiofs",
		"location: /tmp/crux-test",
		"mountPoint: /mnt/crux",
		"writable: true",
		"containerd:",
		"system: false",
		"user: false",
		"containerd.sock",
		"apk add --no-cache containerd",
	}
	for _, s := range required {
		if !strings.Contains(output, s) {
			t.Errorf("config missing %q\nfull output:\n%s", s, output)
		}
	}
}

func TestConfigTemplate_x86(t *testing.T) {
	data := configData{
		Arch:    "x86_64",
		CPUs:    4,
		Memory:  "4GiB",
		Disk:    "20GiB",
		DataDir: "/home/user/.local/share/crux",
	}

	var buf bytes.Buffer
	if err := configTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template execution: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "arch: x86_64") {
		t.Errorf("expected arch x86_64 in output:\n%s", output)
	}
	if !strings.Contains(output, "cpus: 4") {
		t.Errorf("expected cpus 4 in output:\n%s", output)
	}
}
