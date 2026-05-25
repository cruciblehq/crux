//go:build darwin || linux

package local

import (
	"fmt"
	"os"
	"os/user"
	"text/template"

	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/paths"
)

const (

	// Fixed resource allocation for the local shared VM.
	//
	// The local backend runs all services inside a single Lima VM. VM sizing
	// is NOT derived from service affordance declarations — a service
	// requesting 8 CPUs on a 4-core host cannot be satisfied, and launching
	// a separate VM per service would exhaust host resources. Instead these
	// fixed values act as a reasonable developer-workstation baseline. Cloud
	// providers will derive instance sizing from affordances at provisioning
	// time; the local backend intentionally ignores them.
	defaultLimaCPUs      = 2  // Virtual CPUs allocated to the local VM.
	defaultLimaMemoryGiB = 2  // Memory in GiB allocated to the local VM.
	defaultLimaDiskGiB   = 10 // Disk size in GiB allocated to the local VM.

	// containerd socket path inside the Lima VM (guest).
	guestContainerdSocket = "/run/containerd/containerd.sock"
)

// Values injected into the Lima YAML template.
type limaConfig struct {
	Arch        string // Lima architecture identifier (e.g. "aarch64", "x86_64").
	CPUs        int    // Number of virtual CPUs.
	Memory      string // Memory allocation with unit suffix (e.g. "2GiB").
	Disk        string // Disk size with unit suffix (e.g. "10GiB").
	Home        string // Host home directory for the VM mount.
	User        string // Host username (Lima creates a matching guest user).
	UserUID     int    // Host user's numeric UID; baked into the template so the provision script needs no id(1) lookup.
	ImagePath   string // Local path to the cached machine disk image.
	GuestSocket string // containerd socket path inside the VM (guest-local, under /run).
	HostSocket  string // containerd socket path on the host (Lima forwards guest to host).
}

// Lima YAML configuration template, declared in platform-specific files via
// //go:embed so each platform can use a different template file.
var limaConfigTemplate *template.Template

// Generates the Lima YAML configuration for the shared crux VM.
//
// The configuration targets the host's native architecture and uses sensible
// defaults for CPU, memory, and disk allocation. The VM boots from the
// provided machine disk image. containerd runs as a system service inside the
// VM; Lima's portForwards section tunnels the guest containerd socket to the
// host so crux can dial it transparently.
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
		UserUID:     os.Getuid(),
		ImagePath:   imagePath,
		GuestSocket: guestContainerdSocket,
		HostSocket:  paths.ContainerdSocket(limaInstanceName),
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
