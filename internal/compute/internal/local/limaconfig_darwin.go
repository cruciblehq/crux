//go:build darwin

package local

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"text/template"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/compute/internal/provider"
	"github.com/cruciblehq/crux/internal/paths"
)

const (

	// Resource constraints for the Lima VM
	defaultLimaCPUs      = 2  // Default number of virtual CPUs allocated to the VM.
	defaultLimaMemoryGiB = 2  // Default memory in GiB allocated to the VM.
	defaultLimaDiskGiB   = 10 // Default disk size in GiB allocated to the VM.

	// Default GID for the containerd group inside the VM. The containerd
	// socket is configured with this group so the Lima user can access it.
	// Alpine reserves GID 999 for the ping group, so we use 990.
	defaultLimaContainerdGID = 990
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
	Arch             string // Lima architecture identifier (e.g. "aarch64", "x86_64").
	CPUs             int    // Number of virtual CPUs.
	Memory           string // Memory allocation with unit suffix (e.g. "2GiB").
	Disk             string // Disk size with unit suffix (e.g. "10GiB").
	Home             string // Host home directory for the virtiofs mount.
	User             string // Host username (Lima creates a matching guest user).
	ContainerdGID    int    // GID for the containerd group (controls socket access).
	CruxdDownloadURL string // URL to download the cruxd binary from.
	ImagePath        string // Local path to the cached Alpine qcow2 image.
}

// Generates the Lima YAML configuration for the shared crux VM.
//
// The configuration targets the host's native architecture and uses sensible
// defaults for CPU, memory, and disk allocation. The VM boots from the
// provided Alpine Linux image. The cruxd binary is installed during
// provisioning; individual cruxd instances are started on demand, not at
// VM creation time.
func generateLimaConfig(cruxdVersion string, imagePath string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", crex.Wrap(ErrRuntimeConfig, err)
	}

	data := limaConfig{
		Arch:             limaArch(),
		CPUs:             defaultLimaCPUs,
		Memory:           fmt.Sprintf("%dGiB", defaultLimaMemoryGiB),
		Disk:             fmt.Sprintf("%dGiB", defaultLimaDiskGiB),
		Home:             home,
		User:             os.Getenv("USER"),
		ContainerdGID:    defaultLimaContainerdGID,
		CruxdDownloadURL: provider.CruxdDownloadURL(cruxdVersion, runtime.GOARCH),
		ImagePath:        imagePath,
	}

	if err := os.MkdirAll(paths.VMDir(), paths.DefaultDirMode); err != nil {
		return "", crex.Wrap(ErrRuntimeConfig, err)
	}

	configPath := paths.LimaConfig()
	f, err := os.Create(configPath)
	if err != nil {
		return "", crex.Wrap(ErrRuntimeConfig, err)
	}
	defer f.Close()

	if err := limaConfigTemplate.Execute(f, data); err != nil {
		return "", crex.Wrap(ErrRuntimeConfig, err)
	}

	return configPath, nil
}
