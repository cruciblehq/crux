//go:build darwin

package local

import (
	_ "embed"
	"fmt"
	"os"
	"os/user"
	"text/template"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/paths"
	specpaths "github.com/cruciblehq/spec/paths"
)

const (

	// Resource constraints for the Lima VM
	defaultLimaCPUs      = 2  // Default number of virtual CPUs allocated to the VM.
	defaultLimaMemoryGiB = 2  // Default memory in GiB allocated to the VM.
	defaultLimaDiskGiB   = 10 // Default disk size in GiB allocated to the VM.
)

//go:embed templates/lima.yaml.tmpl
var limaConfigSource string

// Lima YAML configuration template.
//
// Uses Virtualization.framework (vz) on macOS with virtiofs mounts. The
// provisioning script installs containerd and CNI plugins on first boot.
var limaConfigTemplate = template.Must(template.New("lima").Parse(limaConfigSource))

// Values injected into the Lima YAML template.
type limaConfig struct {
	Arch        string // Lima architecture identifier (e.g. "aarch64", "x86_64").
	CPUs        int    // Number of virtual CPUs.
	Memory      string // Memory allocation with unit suffix (e.g. "2GiB").
	Disk        string // Disk size with unit suffix (e.g. "10GiB").
	Home        string // Host home directory for the virtiofs mount.
	User        string // Host username (Lima creates a matching guest user).
	ImagePath   string // Local path to the cached machine disk image.
	GuestSocket string // cruxd socket path inside the VM (guest-local, under /run).
	HostSocket  string // cruxd socket path on the host (Lima forwards guest → host).
}

// Generates the Lima YAML configuration for the shared crux VM.
//
// The configuration targets the host's native architecture and uses sensible
// defaults for CPU, memory, and disk allocation. The VM boots from the
// provided machine disk image. All required services (cruxd, containerd) are
// pre-installed in the image; only Lima-specific host setup (user group
// membership, socket directory permissions) is performed during provisioning.
func generateLimaConfig(name string, imagePath string) (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", crex.Wrap(ErrHostConfig, err)
	}

	data := limaConfig{
		Arch:        limaArch(),
		CPUs:        defaultLimaCPUs,
		Memory:      fmt.Sprintf("%dGiB", defaultLimaMemoryGiB),
		Disk:        fmt.Sprintf("%dGiB", defaultLimaDiskGiB),
		Home:        u.HomeDir,
		User:        u.Username,
		ImagePath:   imagePath,
		GuestSocket: specpaths.Socket(name),
		HostSocket:  paths.CruxdSocket(name),
	}

	if err := os.MkdirAll(paths.VMDir(), paths.DefaultDirMode); err != nil {
		return "", crex.Wrap(ErrHostConfig, err)
	}

	configPath := paths.LimaConfig()
	f, err := os.Create(configPath)
	if err != nil {
		return "", crex.Wrap(ErrHostConfig, err)
	}
	defer f.Close()

	if err := limaConfigTemplate.Execute(f, data); err != nil {
		return "", crex.Wrap(ErrHostConfig, err)
	}

	return configPath, nil
}
