// Package resource manages the lifecycle of Crucible resources.
//
// Each resource type has its own [Runner] implementation that handles building,
// packaging (Pack), lifecycle management (Start, Stop, Destroy, Exec, Status),
// and pushing (Push). Not every resource type supports every operation;
// unsupported operations return [ErrUnsupported].
//
// [Resolve] reads a manifest and returns the appropriate Runner for it:
//
//	man, r, err := resource.Resolve("crucible.yaml", resource.Options{
//	    DefaultRegistry:  "http://hub.cruciblehq.xyz:8080",
//	    DefaultNamespace: "crucible",
//	})
//
//	result, err := r.Build(ctx, *man, "build")
//	err = r.Start(ctx, *man, "build/image.tar")
//	err = r.Push(ctx, *man, "dist/package.tar.zst")
package resource
