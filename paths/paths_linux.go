//go:build linux

package paths

import "path/filepath"

// Path to the vendored Lima installation directory.
//
//	~/.local/share/crux/lima
func LimaDir() string {
	return filepath.Join(DataDir(), "lima")
}

// Path to the vendored limactl binary.
//
//	~/.local/share/crux/lima/bin/limactl
func LimactlBin() string {
	return filepath.Join(LimaDir(), "bin", "limactl")
}

// Path to the Lima YAML configuration file for the shared crux VM.
//
//	~/.local/share/crux/vm/lima.yaml
func LimaConfig() string {
	return filepath.Join(VMDir(), "lima.yaml")
}

// Path to the containerd Unix socket for an instance.
//
// On Linux, containerd runs inside a Lima VM. Lima's portForwards tunnels
// the guest socket to this host path over SSH, so the host dials this path
// transparently.
//
//	~/.cache/crux/instances/<name>/containerd.sock
func ContainerdSocket(name string) string {
	return filepath.Join(CacheDir(), "instances", name, "containerd.sock")
}

