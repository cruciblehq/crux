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

// Path to the cruxd Unix socket for an instance.
//
// On Darwin, cruxd binds a guest-local socket inside the Lima VM. Lima's
// portForwards tunnels the guest socket to this host path over SSH, so the
// host dials this path transparently.
//
//	~/Library/Caches/cruxd/instances/<name>/cruxd.sock
func CruxdSocket(name string) string {
	return filepath.Join(xdg.CacheHome, "cruxd", "instances", name, "cruxd.sock")
}

// Path to the cruxd PID file for an instance.
//
// Returns a path alongside the socket on the shared virtiofs mount so the
// host can read the guest PID directly.
//
//	~/Library/Caches/cruxd/instances/<name>/cruxd.pid
func CruxdPIDFile(name string) string {
	return filepath.Join(xdg.CacheHome, "cruxd", "instances", name, "cruxd.pid")
}
