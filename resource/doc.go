// Package resource handles artifact operations for Crucible resources.
//
// The package is organized in two layers. The manifest layer exposes four
// operations that work from a manifest: [Build] compiles a resource, [Verify]
// checks a build directory, [Pack] produces an archive, and [Push] uploads it
// to the registry. [Build] unwraps the typed config from the manifest,
// dispatches to the matching type builder, and writes the resolved manifest
// alongside the artifacts. Some resource types, such as machine, currently
// participate only in packaging and registry operations and do not have a
// builder yet. Container lifecycle operations (start, stop, exec, and so on)
// go through the agent on the machine, not through this package.
//
// The type builder layer lives in subpackages, one per resource type, each
// taking a typed config rather than a manifest: [runtime], [service], [widget],
// and [blueprint]. The runtime and service builders both compose the shared
// [recipe] engine, which in turn uses [oci] to write image tars. Resource
// types without a builder stay in the manifest layer only. Callers with a
// typed config may use a type builder directly, skipping the manifest layer.
//
// Callers read the manifest themselves and pass it to each operation:
//
//	src, err := registry.NewSource("http://hub.cruciblehq.xyz:8080", "acme", dirs.Registry())
//	man, err := resource.ReadManifest(".")
//
//	result, err := resource.Build(ctx, *man, src, ".", "", resource.BuildDir("."))
//	packed, err := resource.Pack(ctx, result.Output, resource.PackagePath("."))
//	archive, err := os.Open(packed.Output)
//	err = resource.Push(ctx, src, *man, archive)
package resource
