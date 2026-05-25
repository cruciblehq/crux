// Package resource handles artifact operations for Crucible resources.
//
// Each resource type has its own [Handler] implementation that handles building
// (Build), packaging (Pack), and pushing (Push). Lifecycle operations (start,
// stop, exec, etc.) are handled directly via the runtime.
//
// [ResolveHandler] selects the appropriate [Handler] for a resource type.
// Callers read the manifest themselves and pass the resource type:
//
//	src := source.New("http://hub.cruciblehq.xyz:8080")
//	man, err := manifest.Read("crucible.yaml")
//	h, err := resource.ResolveHandler(man.Resource.Type, src, filepath.Dir("crucible.yaml"), "")
//
//	result, err := h.Build(ctx, *man, "build")
//	packed, err := h.Pack(ctx, result.Output, "dist/package.tar.zst")
//	err = h.Push(ctx, *man, packed.Output)
package resource
