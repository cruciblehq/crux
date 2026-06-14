// Package files provides Crucible-specific paths and file utilities.
//
// It covers three concerns.
//
// Paths: XDG-based directory and file path resolution for all Crucible data,
// cache, config, and project locations. All path functions are pure string
// operations with no side effects. Project-relative functions accept a base
// directory; system-level functions derive their roots from XDG or OS defaults.
// ValidateAbsPath validates slash-separated absolute paths such as those that
// target locations inside a sandbox.
//
//	fmt.Println(files.Manifest("."))      // crucible.yaml
//	fmt.Println(files.BuildDir("."))      // build
//	fmt.Println(files.DataDir())          // ~/Library/Application Support/crux
//	fmt.Println(files.TempDir())          // /tmp/crux-<uid>
//
// File helpers: atomic writes with SHA-256 digest tracking, existence checks,
// empty-directory pruning, subdirectory enumeration, and managed temporary
// file creation. All helpers operate on plain [os.File] values and filesystem
// paths; they carry no domain knowledge about manifests, caches, or any other
// Crucible concept.
//
// File locking: exclusive locks via flock(2), with context-aware acquisition
// that honours cancellation and deadlines.
package files
