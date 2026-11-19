// Package dirs locates the directories crux keeps on the user's computer.
//
// This is the only place that names them. All of them follow the XDG base
// directory layout and live under the user's home, so no other user can
// create or replace them. What goes inside each directory is up to the
// package that owns it.
package dirs

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

// Directory names.
const (
	app          = "crux"     // Subdirectory under each XDG base directory.
	localName    = "local"    // Local environment blueprint state.
	limaName     = "lima"     // Vendored Lima installation.
	vmName       = "vm"       // Lima home holding the VM.
	registryName = "registry" // Registry cache.
	tempName     = "tmp"      // Temporary staging files.
	trustName    = "trust"    // Certificate authority certificates and revocation list.
	machinesName = "machines" // Machine identifiers and enrollment tokens.
)

// Returns the directory for data crux keeps across runs.
func data() string {
	return filepath.Join(xdg.DataHome, app)
}

// Returns the directory for data crux can download or rebuild again.
func cache() string {
	return filepath.Join(xdg.CacheHome, app)
}

// Returns the directory for the local environment blueprint state.
func Local() string {
	return filepath.Join(data(), localName)
}

// Returns the directory of the vendored Lima installation.
func Lima() string {
	return filepath.Join(data(), limaName)
}

// Returns the Lima home directory that holds the VM.
func VM() string {
	return filepath.Join(data(), vmName)
}

// Returns the directory for the registry cache.
func Registry() string {
	return filepath.Join(cache(), registryName)
}

// Returns the directory for the certificate authority's public files.
//
// The private keys are kept in the OS keychain, not here.
func Trust() string {
	return filepath.Join(data(), trustName)
}

// Returns the directory for machine records.
func Machines() string {
	return filepath.Join(data(), machinesName)
}

// Returns the directory for temporary files.
//
// It sits in the cache directory rather than the system temp directory,
// because a fixed name in a directory every user can write to can be taken
// over by another user.
func Temp() string {
	return filepath.Join(cache(), tempName)
}
