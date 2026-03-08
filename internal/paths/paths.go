package paths

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/cruciblehq/spec/manifest"
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

	// Default permission mode for executable files.
	DefaultExecMode os.FileMode = 0755
)

// Path to the build output directory for a resource project.
//
// Contains built artifacts whose contents depend on the resource type. For
// example, widgets produce compiled JavaScript bundles while services produce
// OCI image tarballs.
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

// Path to the OCI image tarball in a resource project's build output.
func ImageTar(base string) string {
	return filepath.Join(BuildDir(base), "image.tar")
}

// Path to the manifest file for a resource project.
func Manifest(base string) string {
	return filepath.Join(base, manifest.ManifestFile)
}

// Path to the directory for persistent application data.
//
//	Linux:   $XDG_DATA_HOME/crux or ~/.local/share/crux
//	macOS:   ~/Library/Application Support/crux
//	Windows: %LOCALAPPDATA%\crux
func DataDir() string {
	return filepath.Join(xdg.DataHome, DefaultClientName)
}

// Path to the directory for user configuration files.
//
//	Linux:   $XDG_CONFIG_HOME/crux or ~/.config/crux
//	macOS:   ~/Library/Application Support/crux
//	Windows: %APPDATA%\crux
func ConfigDir() string {
	return filepath.Join(xdg.ConfigHome, DefaultClientName)
}

// Path to the providers configuration file.
//
// Linux:   $XDG_CONFIG_HOME/crux/providers.yaml
// macOS:   ~/Library/Application Support/crux/providers.yaml
// Windows: %APPDATA%\crux\providers.yaml
func ProvidersConfig() string {
	return filepath.Join(ConfigDir(), "providers.yaml")
}

// Path to the directory for non-essential cached data.
//
//	Linux:   $XDG_CACHE_HOME/crux or ~/.cache/crux
//	macOS:   ~/Library/Caches/crux
//	Windows: %LOCALAPPDATA%\crux\Cache
func CacheDir() string {
	return filepath.Join(xdg.CacheHome, DefaultClientName)
}

// Path to the VM data directory.
//
// Contains the Lima configuration, disk images, and runtime state for the
// Crucible virtual machine.
//
//	macOS:   ~/Library/Application Support/crux/vm
func VMDir() string {
	return filepath.Join(DataDir(), "vm")
}

// Path to the registry cache directory.
//
// Stores downloaded package archives and extracted contents for offline
// access and fast re-installation.
//
//	Linux:   $XDG_CACHE_HOME/crux/registry or ~/.cache/crux/registry
//	macOS:   ~/Library/Caches/crux/registry
func RegistryCacheDir() string {
	return filepath.Join(CacheDir(), "registry")
}
