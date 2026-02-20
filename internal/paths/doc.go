// Package paths provides Crucible-specific directory and file paths.
//
// All paths are derived from XDG base directories on Unix and appropriate
// system directories on Windows and macOS. This package centralizes path
// definitions to ensure consistency across implementations. Project-relative
// functions like BuildDir and Manifest accept a base directory, while
// system-level functions like Data, Config, and Cache return absolute paths
// based on the current platform.
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
//	fmt.Println(paths.Data())   // e.g., ~/Library/Application Support/crux
//	fmt.Println(paths.Cache())  // e.g., ~/Library/Caches/crux
//	fmt.Println(paths.Store())  // e.g., ~/Library/Caches/crux/store
package paths
