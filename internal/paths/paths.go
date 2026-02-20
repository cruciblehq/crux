package paths

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/adrg/xdg"
	specpaths "github.com/cruciblehq/spec/paths"
)

const (

	// The default name for the Crucible client.
	DefaultClientName = "crux"

	// Default permission mode used when creating directories.
	//
	// This mode is required when handling resource extraction and storage and
	// optional for other purposes.
	DefaultDirMode os.FileMode = 0755

	// Default permission mode used when creating files.
	//
	// This mode is required when handling resource extraction and storage and
	// optional for other purposes.
	DefaultFileMode os.FileMode = 0644
)

// Path to the build output directory for a resource project.
//
// Contains built artifacts whose contents depend on the resource type.
// For example, widgets produce compiled JavaScript bundles while services
// produce OCI image tarballs.
func BuildDir(base string) string {
	return filepath.Join(base, "build")
}

// Path to the dist output directory for a resource project.
//
// Contains packaged archives and generated plans.
func DistDir(base string) string {
	return filepath.Join(base, "dist")
}

// Path to the default package archive for a resource project.
func Package(base string) string {
	return filepath.Join(DistDir(base), "package.tar.zst")
}

// Path to the manifest file for a resource project.
func Manifest(base string) string {
	return filepath.Join(base, "crucible.yaml")
}

// Path to the directory for persistent application data.
//
//	Linux:   $XDG_DATA_HOME/crux or ~/.local/share/crux
//	macOS:   ~/Library/Application Support/crux
//	Windows: %LOCALAPPDATA%\crux
func Data() string {
	return filepath.Join(xdg.DataHome, DefaultClientName)
}

// Path to the directory for user configuration files.
//
//	Linux:   $XDG_CONFIG_HOME/crux or ~/.config/crux
//	macOS:   ~/Library/Application Support/crux
//	Windows: %APPDATA%\crux
func Config() string {
	return filepath.Join(xdg.ConfigHome, DefaultClientName)
}

// Path to the providers configuration file.
//
// Linux:   $XDG_CONFIG_HOME/crux/providers.yaml
// macOS:   ~/Library/Application Support/crux/providers.yaml
// Windows: %APPDATA%\crux\providers.yaml
func Providers() string {
	return filepath.Join(Config(), "providers.yaml")
}

// Path to the directory for non-essential cached data.
//
//	Linux:   $XDG_CACHE_HOME/crux or ~/.cache/crux
//	macOS:   ~/Library/Caches/crux
//	Windows: %LOCALAPPDATA%\crux\Cache
func Cache() string {
	return filepath.Join(xdg.CacheHome, DefaultClientName)
}

// Path to the store cache directory.
//
//	Linux:   $XDG_CACHE_HOME/crux/store or ~/.cache/crux/store
//	macOS:   ~/Library/Caches/crux/store
//	Windows: %LOCALAPPDATA%\crux\Cache\store
func Store() string {
	return filepath.Join(Cache(), "store")
}

// Path to the VM data directory.
//
// Contains the Lima configuration, disk images, and runtime state for the
// Crucible virtual machine.
//
//	macOS:   ~/Library/Application Support/crux/vm
func VM() string {
	return filepath.Join(Data(), "vm")
}

// Path to the cruxd daemon Unix socket.
//
// On Linux this returns the canonical system path defined by the spec
// package. On macOS (development) it returns a host-local path where
// Lima forwards the guest socket.
//
//	Linux:   /run/cruxd/cruxd.sock
//	macOS:   ~/Library/Caches/cruxd/run/cruxd.sock
func DaemonSocket() string {
	if runtime.GOOS == "linux" {
		return specpaths.Socket()
	}
	// macOS: host-side path for the Lima-forwarded socket.
	return filepath.Join(xdg.CacheHome, "cruxd", "run", "cruxd.sock")
}
