package paths

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/adrg/xdg"
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

// Path to the directory for executable binaries.
//
//	Linux:   $XDG_BIN_HOME or ~/.local/bin
//	macOS:   ~/Library/Application Support/crux/bin
//	Windows: %LOCALAPPDATA%\crux\bin
func Bin() string {
	return filepath.Join(xdg.BinHome, DefaultClientName)
}

// Path to the store cache directory.
//
// Contains cached resources fetched from remote registries.
//
//	Linux:   $XDG_CACHE_HOME/crux/store or ~/.cache/crux/store
//	macOS:   ~/Library/Caches/crux/store
//	Windows: %LOCALAPPDATA%\crux\Cache\store
func Store() string {
	return filepath.Join(Cache(), "store")
}

// Path to the directory for runtime files (sockets, PIDs).
//
//	Linux:   $XDG_RUNTIME_DIR/crux or /run/user/<uid>/crux
//	macOS:   ~/Library/Caches/crux/run
//	Windows: %LOCALAPPDATA%\crux\run
func Runtime() string {
	if xdg.RuntimeDir != "" {
		return filepath.Join(xdg.RuntimeDir, DefaultClientName)
	}
	// Fallback for macOS/Windows
	return filepath.Join(Cache(), "run")
}

// Path to the directory for log files.
//
//	Linux:   $XDG_STATE_HOME/crux/logs or ~/.local/state/crux/logs
//	macOS:   ~/Library/Logs/crux
//	Windows: %LOCALAPPDATA%\crux\logs
func Logs() string {
	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Logs", DefaultClientName)
	}
	// Linux: XDG_STATE_HOME
	if xdg.StateHome != "" {
		return filepath.Join(xdg.StateHome, DefaultClientName, "logs")
	}
	// Fallback
	return filepath.Join(Data(), "logs")
}

// Path to the server log file.
//
//	Linux:   ~/.local/state/crux/logs/server.log
//	macOS:   ~/Library/Logs/crux/server.log
//	Windows: %LOCALAPPDATA%\crux\logs\server.log
func ServerLog() string {
	return filepath.Join(Logs(), "server.log")
}
