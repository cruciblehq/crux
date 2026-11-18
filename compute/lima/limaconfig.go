//go:build darwin || linux

package lima

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/cruciblehq/crux/dirs"
	speccompute "github.com/cruciblehq/spec/compute"
	"github.com/cruciblehq/spec/trust"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

const (

	// Fixed resource allocation for the local shared VM.
	//
	// The local provider runs all services inside a single Lima VM, so VM sizing
	// is not derived from grants. Instead, these fixed values act as a baseline.
	// Cloud providers will derive machine sizing from grants at provisioning
	// time; the local provider intentionally (and silently) ignores them.
	defaultLimaCPUs      = 2  // Virtual CPUs allocated to the local VM.
	defaultLimaMemoryGiB = 2  // Memory in GiB allocated to the local VM.
	defaultLimaDiskGiB   = 10 // Disk size in GiB allocated to the local VM.

	// Mount point of the machine resource directory inside the Lima VM.
	guestMachineDir = "/mnt/crucible/machine"

	// Prefix of the cruxd binary file names shipped with the machine resource.
	cruxdBinaryPrefix = "cruxd-"

	// Lima configuration file in the VM directory.
	limaConfigFile = "lima.yaml"
)

// Path to the Lima YAML configuration file for the shared crux VM.
func limaConfigPath() string {
	return filepath.Join(dirs.VM(), limaConfigFile)
}

// Lima YAML configuration template.
//
// This is declared in platform-specific files via "go:embed" so each platform
// can use a different template file.
var limaConfigTemplate *template.Template

// Values injected into the Lima YAML template.
//
// Provision values (CPUs, Memory, and Disk) are fixed since the local provider
// does not derive VM sizing from grants. These values are passed directly to
// the hypervisor without being validated against host resources, so these they
// constitute minimum requirements (defaultLimaCPUs, defaultLimaMemoryGiB, and
// defaultLimaDiskGiB). ImagePath is the local path to the machine disk
// image, which is downloaded from the Crucible registry and cached locally if
// not already cached. The machine directory holding the image also holds the
// cruxd binary; it is mounted read-only into the VM so the provision script
// can install and start cruxd, whose port is forwarded to the host. Enrollment
// is written into the VM at EnrollmentPath on every boot; cruxd deletes it
// once enrolled. The config file is readable only by the user, because it
// holds the enrollment token.
type limaConfig struct {
	Arch            string // Lima architecture identifier ("aarch64" or "x86_64").
	CPUs            int    // Number of virtual CPUs (defaultLimaCPUs).
	Memory          string // Memory allocation (defaultLimaMemoryGiB with "GiB" suffix).
	Disk            string // Disk size (defaultLimaDiskGiB with "GiB" suffix).
	ImagePath       string // Local path to the cached machine disk image.
	MachineDir      string // Host directory of the extracted machine resource.
	GuestMachineDir string // Mount point of MachineDir inside the VM.
	CruxdBinary     string // Path inside the VM to the cruxd binary for the VM's architecture.
	CruxdPort       int    // Port cruxd listens on, forwarded to the same port on the host.
	EnrollmentPath  string // Path inside the VM where cruxd reads its enrollment data.
	Enrollment      string // Enrollment data as a single line of JSON.
}

// Builds the Lima YAML configuration template data.
//
// Targets the host's native architecture and uses defaults for CPU, memory,
// and disk allocation. The VM boots from the machine disk image at imagePath.
// containerd runs as a system service inside the VM. cruxd is installed from
// the directory holding imagePath and its port is forwarded to the host.
// enrollment is the machine's enrollment data. Does not touch the filesystem.
func buildLimaConfig(imagePath string, enrollment []byte) limaConfig {
	return limaConfig{
		Arch:            limaArch(),
		CPUs:            defaultLimaCPUs,
		Memory:          fmt.Sprintf("%dGiB", defaultLimaMemoryGiB),
		Disk:            fmt.Sprintf("%dGiB", defaultLimaDiskGiB),
		ImagePath:       imagePath,
		MachineDir:      filepath.Dir(imagePath),
		GuestMachineDir: guestMachineDir,
		CruxdBinary:     path.Join(guestMachineDir, fmt.Sprintf("%s%s", cruxdBinaryPrefix, limaArch())),
		CruxdPort:       speccompute.DefaultPort,
		EnrollmentPath:  trust.EnrollmentPath,
		Enrollment:      string(enrollment),
	}
}

// Writes the Lima YAML configuration file and returns its path.
//
// The configuration is generated with [buildLimaConfig] and rendered using
// [limaConfigTemplate]. The file is written to disk so it can be read by
// limactl when provisioning the VM, and is readable only by the user. If the
// file already exists it will be overwritten. Returns the path to the file
// that was generated.
func generateLimaConfig(imagePath string, enrollment []byte) (string, error) {
	const description = "cannot configure local environment"

	data := buildLimaConfig(imagePath, enrollment)

	configPath := limaConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), file.DefaultDirMode); err != nil {
		return "", crex.SystemError(description, "failed to create the configuration directory").
			Recoveryf("Make sure you have write access to %s, then try again.", filepath.Dir(configPath)).
			Cause(crex.Wrap(ErrMachineConfig, err)).
			Err()
	}

	f, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.OwnerFileMode)
	if err != nil {
		return "", crex.SystemError(description, "failed to write the configuration file").
			Recoveryf("Make sure you have write access to %s, then try again.", configPath).
			Cause(crex.Wrap(ErrMachineConfig, err)).
			Err()
	}
	defer f.Close()
	if err := f.Chmod(file.OwnerFileMode); err != nil {
		return "", crex.SystemError(description, "failed to restrict access to the configuration file").
			Recoveryf("Make sure you own %s, then try again.", configPath).
			Cause(crex.Wrap(ErrMachineConfig, err)).
			Err()
	}

	if err := limaConfigTemplate.Execute(f, data); err != nil {
		return "", crex.SystemError(description, "failed to render the configuration file").
			Recovery("If the problem persists, report it to the Crucible team.").
			Cause(crex.Wrap(ErrMachineConfig, err)).
			Err()
	}

	return configPath, nil
}
