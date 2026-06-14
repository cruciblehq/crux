package files

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/adrg/xdg"
	"github.com/cruciblehq/crux/crex"
)

// Default permission modes.
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
)

// Default name for the Crucible client.
const DefaultClientName = "crux"

// Standard filenames.
const (

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

	// Standard filename for the Lima VM configuration.
	LimaConfigFile = "lima.yaml"

	// Standard filename for the containerd Unix socket.
	ContainerdSocketFile = "containerd.sock"
)

// Subdirectory names.
const (

	// Subdirectory name for build artifacts within a project directory.
	BuildDirName = "build"

	// Subdirectory name for distributable archives within a project directory.
	DistDirName = "dist"

	// Subdirectory name for VM data within the application data directory.
	VMDirName = "vm"

	// Subdirectory name for the registry cache within the cache directory.
	RegistryDirName = "registry"

	// Subdirectory name for extracted resource contents within the registry cache.
	RegistryExtractedDirName = "extracted"

	// Subdirectory name for the local blueprint within the application data directory.
	LocalDirName = "local"

	// Subdirectory name for the Lima installation within the application data directory.
	LimaDirName = "lima"

	// Subdirectory name for instance sockets within the cache directory.
	InstancesDirName = "instances"
)

// Name of the limactl binary.
const LimactlBinName = "limactl"

// Permission mode for the temp directory, reachable only by the current user.
const tempDirMode os.FileMode = 0700

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

// Path to the directory for extracted contents of a cached registry resource.
//
// Locates the unpacked archive for the given namespace, resource, and version
// under the registry cache extracted subtree.
func RegistryExtractedVersionDir(namespace, resource, version string) string {
	return filepath.Join(RegistryCacheDir(), RegistryExtractedDirName, namespace, resource, version)
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

// Path to the directory for temporary Crucible files.
//
// Each user gets their own directory, reachable only by them, so files placed
// in the shared OS temp dir cannot be read or hijacked by anyone else. On Unix
// the user id is added to the name so paths do not collide between users.
//
//	Linux/macOS:  /tmp/crux-<uid>
//	Windows:      %TEMP%\crux
func TempDir() string {
	name := DefaultClientName
	if uid := os.Getuid(); uid >= 0 {
		name += "-" + strconv.Itoa(uid)
	}
	return filepath.Join(os.TempDir(), name)
}

// Creates a new temporary file in the Crucible temp directory.
//
// The directory is created if missing, reachable only by the current user. The
// pattern follows [os.CreateTemp]: a * is replaced with a random string.
// Callers must close and remove the file when done.
func CreateTemp(pattern string) (*os.File, error) {
	dir := TempDir()
	if err := os.MkdirAll(dir, tempDirMode); err != nil {
		return nil, err
	}
	if err := verifyTempDir(dir); err != nil {
		return nil, err
	}
	return os.CreateTemp(dir, pattern)
}

// Verifies that dir is a real directory rather than a symlink.
//
// Guards against hijacking in the shared OS temp dir by rejecting anything
// that is a symlink or not a directory, and tightening permissions when the
// directory already existed with looser access.
func verifyTempDir(dir string) error {
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return crex.Wrapf(ErrUnsafeTempDir, "%q is a symlink or not a directory", dir)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return os.Chmod(dir, tempDirMode)
	}
	return nil
}
