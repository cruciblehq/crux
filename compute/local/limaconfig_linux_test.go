//go:build linux

package local

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfigTemplate_IsValid(t *testing.T) {
	data := limaConfig{
		Arch:            "aarch64",
		CPUs:            2,
		Memory:          "2GiB",
		Disk:            "10GiB",
		ImagePath:       "/tmp/vm/machine.qcow2",
		MachineDir:      "/tmp/vm",
		GuestMachineDir: "/mnt/crucible/machine",
		CruxdBinary:     "/mnt/crucible/machine/cruxd-aarch64",
		CruxdPort:       2789,
		CruxdHealthPath: "/health",
	}

	var buf bytes.Buffer
	if err := limaConfigTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template execution: %v", err)
	}

	output := buf.String()
	required := []string{
		"vmType: qemu",
		"arch: aarch64",
		"cpus: 2",
		"memory: 2GiB",
		"disk: 10GiB",
		"mountType: 9p",
		"containerd:",
		"system: false",
		"user: false",
		"portForwards:",
		`location: "/tmp/vm"`,
		`mountPoint: "/mnt/crucible/machine"`,
		"writable: false",
		`install -m 0755 "/mnt/crucible/machine/cruxd-aarch64" /usr/local/bin/cruxd`,
		`wget -q -O /dev/null "http://127.0.0.1:2789/health"`,
		"guestPort: 2789",
		"hostPort: 2789",
	}
	for _, s := range required {
		if !strings.Contains(output, s) {
			t.Errorf("config missing %q\nfull output:\n%s", s, output)
		}
	}
}
