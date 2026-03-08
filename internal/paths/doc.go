// Package paths provides Crucible-specific directory and file paths.
//
// All paths are derived from XDG base directories on Unix and appropriate
// system directories on Windows and macOS. This package centralizes path
// definitions to ensure consistency across implementations. Project-relative
// functions like BuildDir and Manifest accept a base directory, while
// system-level functions like DataDir, ConfigDir, and CacheDir return
// absolute paths based on the current platform.
//
// Resolving project paths from a working directory:
//
//	base := "."
//	fmt.Println(paths.Manifest(base))  // crucible.yaml
//	fmt.Println(paths.BuildDir(base))  // build
//	fmt.Println(paths.Package(base))   // dist/package.tar.zst
//
// Locating system directories for data and cache storage:
//
//	fmt.Println(paths.DataDir())   // e.g., ~/Library/Application Support/crux
//	fmt.Println(paths.CacheDir())  // e.g., ~/Library/Caches/crux
package paths
