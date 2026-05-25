package paths

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

const (

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

	// Default name for the Crucible client.
	DefaultClientName = "crux"

	// Standard filename for Crucible manifests.
	ManifestFile = "crucible.yaml"

	// Standard filename for OCI image tarballs produced by runtime and service builds.
	ImageFile = "image.tar"

	// Standard filename for the JavaScript bundle produced by widget builds.
	WidgetMainFile = "index.js"

	// Standard filename for the deployment plan produced by blueprint builds.
	PlanFile = "plan.yaml"

	// Standard filename for the packaged resource archive.
	PackageFile = "package.tar.zst"

	// Standard filename for the providers configuration file.
	ProvidersFile = "providers.yaml"

	// Subdirectory name for build artifacts within a project directory.
	BuildDirName = "build"

	// Subdirectory name for distributable archives within a project directory.
	DistDirName = "dist"

	// Subdirectory name for VM data within the application data directory.
	VMDirName = "vm"

	// Subdirectory name for the registry cache within the cache directory.
	RegistryDirName = "registry"

	// Subdirectory name for the local blueprint within the application data directory.
	LocalDirName = "local"

	// Subdirectory name for the Lima installation within the application data directory.
	LimaDirName = "lima"

	// Name of the limactl binary.
	LimactlBinName = "limactl"

	// Standard filename for the Lima VM configuration.
	LimaConfigFile = "lima.yaml"

	// Subdirectory name for instance sockets within the cache directory.
	InstancesDirName = "instances"

	// Standard filename for the containerd Unix socket.
	ContainerdSocketFile = "containerd.sock"
)

// Path to the build output directory for a resource project.
//
// Contains built artifacts whose contents depend on the resource type. For
// example, widgets produce compiled JavaScript bundles while services produce
// OCI image tarballs.
func BuildDir(base string) string {
	return filepath.Join(base, BuildDirName)
}

// Path to the dist output directory for a resource project.
//
// Contains packaged archives and generated plans.
func DistDir(base string) string {
	return filepath.Join(base, DistDirName)
}

// Path to the default package archive for a resource project.
func Package(base string) string {
	return filepath.Join(DistDir(base), PackageFile)
}

// Path to the OCI image tarball in a resource project's build output.
func BuildImage(base string) string {
	return filepath.Join(BuildDir(base), ImageFile)
}

// Path to the manifest file for a resource project.
func Manifest(base string) string {
	return filepath.Join(base, ManifestFile)
}

// Path to the deployment plan file for a resource project.
func Plan(base string) string {
	return filepath.Join(base, PlanFile)
}

// Path to the JavaScript bundle entry point for a resource project.
func WidgetMain(base string) string {
	return filepath.Join(base, WidgetMainFile)
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
//	Linux:   $XDG_CONFIG_HOME/crux/providers.yaml
//	macOS:   ~/Library/Application Support/crux/providers.yaml
//	Windows: %APPDATA%\crux\providers.yaml
func ProvidersConfig() string {
	return filepath.Join(ConfigDir(), ProvidersFile)
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
	return filepath.Join(DataDir(), VMDirName)
}

// Path to the registry cache directory.
//
// Stores downloaded package archives and extracted contents for offline
// access and fast re-installation.
//
//	Linux:   $XDG_CACHE_HOME/crux/registry or ~/.cache/crux/registry
//	macOS:   ~/Library/Caches/crux/registry
func RegistryCacheDir() string {
	return filepath.Join(CacheDir(), RegistryDirName)
}

// Path to the default local blueprint directory.
//
// Contains the crucible.yaml manifest that defines which services are
// registered in the local environment.
//
//	Linux:   $XDG_DATA_HOME/crux/local or ~/.local/share/crux/local
//	macOS:   ~/Library/Application Support/crux/local
func LocalDir() string {
	return filepath.Join(DataDir(), LocalDirName)
}
