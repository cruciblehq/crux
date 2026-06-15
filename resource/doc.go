// Package resource handles artifact operations for Crucible resources.
//
// The package is organized in two layers. The manifest layer exposes four
// operations keyed off the manifest resource type: [Build] compiles a resource,
// [Verify] checks a build directory, [Pack] produces a distributable archive,
// and [Push] uploads it to the registry. [Build] unwraps the typed config from
// the manifest, dispatches to the matching type builder, and writes the
// resolved manifest alongside the artifacts. Lifecycle operations (start, stop,
// exec, etc.) are handled directly via the runtime.
//
// The type builder layer lives in subpackages, one per resource type, each
// taking a typed config rather than a manifest: [runtime], [service], [widget],
// [affordance], and [blueprint]. The runtime and service builders both compose
// the shared [recipe] engine, which in turn uses [oci] to write image tars.
// Callers with a typed config may use a type builder directly, skipping the
// manifest layer.
//
// Callers read the manifest themselves and pass it to each operation:
//
//	src, err := registry.NewSource("http://hub.cruciblehq.xyz:8080", "acme")
//	man, err := manifest.Read("crucible.yaml")
//
//	result, err := resource.Build(ctx, *man, src, filepath.Dir("crucible.yaml"), "", "build")
//	packed, err := resource.Pack(ctx, result.Output, "dist/package.tar.zst")
//	err = resource.Push(ctx, src, *man, packed.Output)
package resource
