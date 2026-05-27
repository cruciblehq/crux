//go:build darwin || linux

package local

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"text/template"

	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/paths"
	"github.com/cruciblehq/crux/security"
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

	// containerd socket path inside the Lima VM (guest).
	guestContainerdSocket = "/run/containerd/containerd.sock"
)

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
// defaultLimaDiskGiB). Sysctls and nftables rules are injected from the compute
// policy, allowing the executor to configure VM-level security for services
// running inside the local VM. ImagePath is the local path to the machine disk
// image, which is downloaded from the Crucible registry and cached locally if
// not already cached. The guest and host socket paths are used to set up Lima
// port forwarding so crux can dial containerd inside the VM.
type limaConfig struct {
	Arch        string               // Lima architecture identifier ("aarch64" or "x86_64").
	CPUs        int                  // Number of virtual CPUs (defaultLimaCPUs).
	Memory      string               // Memory allocation (defaultLimaMemoryGiB with "GiB" suffix).
	Disk        string               // Disk size (defaultLimaDiskGiB with "GiB" suffix).
	User        string               // Host username (Lima creates a matching guest user).
	UserUID     int                  // Host user's numeric UID.
	ImagePath   string               // Local path to the cached machine disk image.
	GuestSocket string               // containerd socket path inside the VM (guest-local, under /run).
	HostSocket  string               // containerd socket path on the host (Lima forwards guest to host).
	Sysctls     map[string]string    // Boot-time sysctl values from the compute policy.
	NftRules    []security.VMNftRule // VM-level nftables rules from the compute policy.
}

// Builds the Lima YAML configuration template data.
//
// Targets the host's native architecture and uses defaults for CPU, memory,
// and disk allocation. The user is set to the current host username and UID,
// which Lima uses to create a matching guest user. The VM boots from the
// machine disk image at imagePath. containerd runs as a system service inside
// the VM; Lima's portForwards section tunnels the guest socket to the host so
// crux can dial it. If policy is non-nil, its sysctls and nftables rules are
// injected into the config. Does not touch the filesystem.
func buildLimaConfig(imagePath string, policy *manifest.ComputePolicy) (limaConfig, error) {
	u, err := user.Current()
	if err != nil {
		return limaConfig{}, crex.Wrap(ErrHostConfig, err)
	}

	data := limaConfig{
		Arch:        limaArch(),
		CPUs:        defaultLimaCPUs,
		Memory:      fmt.Sprintf("%dGiB", defaultLimaMemoryGiB),
		Disk:        fmt.Sprintf("%dGiB", defaultLimaDiskGiB),
		User:        u.Username,
		UserUID:     os.Getuid(),
		ImagePath:   imagePath,
		GuestSocket: guestContainerdSocket,
		HostSocket:  paths.ContainerdSocket(limaInstanceName),
	}

	if policy != nil {
		data.Sysctls = policy.VM.Sysctls
		data.NftRules = policy.VM.Nftables
	}

	return data, nil
}

// Writes the Lima YAML configuration file and returns its path.
//
// The configuration is generated with [buildLimaConfig] and rendered using
// [limaConfigTemplate]. The file is written to disk so it can be read by
// limactl when provisioning the VM. If the file already exists it will be
// overwritten. Returns the path to the file that was generated.
func generateLimaConfig(imagePath string, policy *manifest.ComputePolicy) (string, error) {
	data, err := buildLimaConfig(imagePath, policy)
	if err != nil {
		return "", err
	}

	configPath := paths.LimaConfig()
	if err := os.MkdirAll(filepath.Dir(configPath), paths.DefaultDirMode); err != nil {
		return "", crex.Wrap(ErrHostConfig, err)
	}

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
