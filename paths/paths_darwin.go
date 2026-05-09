//go:build darwin

package paths

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

// Path to the vendored Lima installation directory.
//
//	~/Library/Application Support/crux/lima
func LimaDir() string {
	return filepath.Join(DataDir(), "lima")
}

// Path to the vendored limactl binary.
//
//	~/Library/Application Support/crux/lima/bin/limactl
func LimactlBin() string {
	return filepath.Join(LimaDir(), "bin", "limactl")
}

// Path to the Lima YAML configuration file for the shared crux VM.
//
//	~/Library/Application Support/crux/vm/lima.yaml
func LimaConfig() string {
	return filepath.Join(VMDir(), "lima.yaml")
}

// Path to the containerd Unix socket for an instance.
//
// On Darwin, containerd runs inside a Lima VM. Lima's portForwards tunnels
// the guest socket to this host path over SSH, so the host dials this path
// transparently.
//
//	~/Library/Caches/crux/instances/<name>/containerd.sock
func ContainerdSocket(name string) string {
	return filepath.Join(xdg.CacheHome, "crux", "instances", name, "containerd.sock")
}
