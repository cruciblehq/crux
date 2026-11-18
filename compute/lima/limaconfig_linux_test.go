//go:build linux

package lima

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
		EnrollmentPath:  "/etc/cruxd/enrollment.json",
		Enrollment:      `{"machine":"abc"}`,
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
		`cat > "/etc/cruxd/enrollment.json" <<'EOF'`,
		`{"machine":"abc"}`,
		"guestPort: 2789",
		"hostPort: 2789",
	}
	for _, s := range required {
		if !strings.Contains(output, s) {
			t.Errorf("config missing %q\nfull output:\n%s", s, output)
		}
	}
}
