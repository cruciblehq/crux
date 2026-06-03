//go:build darwin || linux

package files

import "path/filepath"

// Path to the vendored Lima installation directory.
//
//	<data>/crux/lima
func LimaDir() string {
	return filepath.Join(DataDir(), LimaDirName)
}

// Path to the vendored limactl binary.
//
//	<data>/crux/lima/bin/limactl
func LimactlBin() string {
	return filepath.Join(LimaDir(), "bin", LimactlBinName)
}

// Path to the Lima YAML configuration file for the shared crux VM.
//
//	<data>/crux/vm/lima.yaml
func LimaConfig() string {
	return filepath.Join(VMDir(), LimaConfigFile)
}

// Path to the containerd Unix socket for an instance.
//
// containerd runs inside a Lima VM. Lima's portForwards tunnels the guest
// socket to this host path over SSH, so the host dials this path.
//
//	<cache>/crux/instances/<name>/containerd.sock
func ContainerdSocket(name string) string {
	return filepath.Join(CacheDir(), InstancesDirName, name, ContainerdSocketFile)
}
