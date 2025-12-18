// Package build provides the core logic for building Crucible resources.
//
// It orchestrates the build lifecycle based on the project manifest, delegating
// to specific handlers for different resource types (Widgets, Services, etc).
//
// The package handles:
//
//   - Build Orchestration: Reads the manifest and dispatches the build process
//     to the appropriate handler (e.g., esbuild for widgets).
//   - Resource Type Support: Handles specific build requirements for different
//     resource types defined in the manifest.
//   - Custom Resolution: Includes an internal esbuild plugin to handle Crucible
//     specific module resolution and externals.
//
// The main entry point is [BuildResource], which accepts [manifest.Manifest]
// options to configure the build.
package build
