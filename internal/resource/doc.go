// Package resource handles artifact operations for Crucible resources.
//
// Each resource type has its own [Builder] implementation that handles building
// (Build), packaging (Pack), and pushing (Push). Lifecycle operations (start,
// stop, exec, etc.) are handled by the provider layer via the daemon client.
//
// [ResolveBuilder] reads a manifest and returns the appropriate Builder for it:
//
//	opts := resource.NewOptions(client, "http://hub.cruciblehq.xyz:8080", "crucible")
//	man, b, err := resource.ResolveBuilder(ctx, "crucible.yaml", opts)
//
//	result, err := b.Build(ctx, *man, "build")
//	packed, err := b.Pack(ctx, result.Output, "dist/package.tar.zst")
//	err = b.Push(ctx, *man, packed.Output)
package resource
