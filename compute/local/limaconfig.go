//go:build darwin || linux

package local

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"text/template"

	speccompute "github.com/cruciblehq/spec/compute"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

const (

	// Fixed resource allocation for the local shared VM.
	//
	// The local backend runs all services inside a single Lima VM, so VM sizing
	// is not derived from grants. Instead, these fixed values act as a baseline.
	// Cloud providers will derive instance sizing from grants at provisioning
	// time; the local backend intentionally (and silently) ignores them.
	defaultLimaCPUs      = 2  // Virtual CPUs allocated to the local VM.
	defaultLimaMemoryGiB = 2  // Memory in GiB allocated to the local VM.
	defaultLimaDiskGiB   = 10 // Disk size in GiB allocated to the local VM.

	// Mount point of the machine resource directory inside the Lima VM.
	guestMachineDir = "/mnt/crucible/machine"

	// Prefix of the cruxd binary file names shipped with the machine resource.
	cruxdBinaryPrefix = "cruxd-"
)

// Path to the Lima YAML configuration file for the shared crux VM.
func limaConfigPath() string {
	return filepath.Join(vmDir(), "lima.yaml")
}

// Lima YAML configuration template.
//
// This is declared in platform-specific files via "go:embed" so each platform
// can use a different template file.
var limaConfigTemplate *template.Template

// Values injected into the Lima YAML template.
//
// Provision values (CPUs, Memory, and Disk) are fixed since the local backend
// does not derive VM sizing from grants. These values are passed directly to
// the hypervisor without being validated against host resources, so these they
// constitute minimum requirements (defaultLimaCPUs, defaultLimaMemoryGiB, and
// defaultLimaDiskGiB). ImagePath is the local path to the machine disk
// image, which is downloaded from the Crucible registry and cached locally if
// not already cached. The machine directory holding the image also holds the
// cruxd binary; it is mounted read-only into the VM so the provision script
// can install and start cruxd, whose port is forwarded to the host.
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
	CruxdHealthPath string // Path of cruxd's health endpoint, used by the readiness probe.
}

// Builds the Lima YAML configuration template data.
//
// Targets the host's native architecture and uses defaults for CPU, memory,
// and disk allocation. The VM boots from the machine disk image at imagePath.
// containerd runs as a system service inside the VM. cruxd is installed from
// the directory holding imagePath and its port is forwarded to the host. Does
// not touch the filesystem.
func buildLimaConfig(imagePath string) limaConfig {
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
		CruxdHealthPath: speccompute.EndpointHealth,
	}
}

// Writes the Lima YAML configuration file and returns its path.
//
// The configuration is generated with [buildLimaConfig] and rendered using
// [limaConfigTemplate]. The file is written to disk so it can be read by
// limactl when provisioning the VM. If the file already exists it will be
// overwritten. Returns the path to the file that was generated.
func generateLimaConfig(imagePath string) (string, error) {
	const description = "cannot configure local environment"

	data := buildLimaConfig(imagePath)

	configPath := limaConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), file.DefaultDirMode); err != nil {
		return "", crex.SystemError(description, "failed to create the configuration directory").
			Recoveryf("Make sure you have write access to %s, then try again.", filepath.Dir(configPath)).
			Cause(crex.Wrap(ErrHostConfig, err)).
			Err()
	}

	f, err := os.Create(configPath)
	if err != nil {
		return "", crex.SystemError(description, "failed to write the configuration file").
			Recoveryf("Make sure you have write access to %s, then try again.", configPath).
			Cause(crex.Wrap(ErrHostConfig, err)).
			Err()
	}
	defer f.Close()

	if err := limaConfigTemplate.Execute(f, data); err != nil {
		return "", crex.SystemError(description, "failed to render the configuration file").
			Recovery("If the problem persists, report it to the Crucible team.").
			Cause(crex.Wrap(ErrHostConfig, err)).
			Err()
	}

	return configPath, nil
}
