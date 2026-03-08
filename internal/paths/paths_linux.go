//go:build linux

package paths

import (
	"path/filepath"

	specpaths "github.com/cruciblehq/spec/paths"
)

// Path to the installed cruxd binary on the host.
//
//	~/.local/share/crux/bin/cruxd
func CruxdBin() string {
	return filepath.Join(DataDir(), "bin", "cruxd")
}

// Path to the cruxd Unix socket for an instance.
//
// Returns the canonical system path defined by the spec package.
//
//	/run/cruxd/instances/<name>/cruxd.sock
func CruxdSocket(name string) string {
	return specpaths.Socket(name)
}

// Path to the cruxd PID file for an instance.
//
// Returns the canonical system path defined by the spec package.
//
//	/run/cruxd/instances/<name>/cruxd.pid
func CruxdPIDFile(name string) string {
	return specpaths.PIDFile(name)
}

// Path to the runtime directory for a cruxd instance.
//
// Contains the Unix socket and PID file while the daemon is running.
// Returns the canonical system path defined by the spec package.
//
//	/run/cruxd/instances/<name>
func CruxdInstanceDir(name string) string {
	return specpaths.InstanceDir(name)
}
