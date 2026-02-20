// Package resource provides workflows for packaging, publishing, and
// downloading Crucible resources.
//
// Pack creates a zstd-compressed tar archive from a manifest and build
// artifacts directory. The archive is ready for deployment or distribution
// to a Crucible Hub. The packaging process validates that the resource
// structure matches its declared type before creating the archive.
//
//	result, err := resource.Pack(ctx, resource.PackOptions{
//	    Manifest: "crucible.yaml",
//	    Dist:     "build",
//	    Output:   "dist/package.tar.zst",
//	})
//
// Push uploads a packaged resource to a Hub registry. It handles namespace
// verification, resource creation, version management, and archive upload.
// After a successful push, the local cache is updated to avoid redundant
// downloads.
//
//	err := resource.Push(ctx, resource.PushOptions{
//	    Registry: "http://hub.cruciblehq.xyz:8080",
//	    Manifestfile:     "crucible.yaml",
//	    Package:          "dist/package.tar.zst",
//	    DefaultNamespace: "crucible",
//	})
//
// Pull downloads a resource archive from a remote registry and stores it in
// the local cache. If the resource is already cached with the correct digest,
// no download occurs. Both version-based and channel-based references are
// supported.
//
//	result, err := resource.Pull(ctx, resource.PullOptions{
//	    Registry:         "http://hub.cruciblehq.xyz:8080",
//	    Reference:        "crucible/login 1.0.0",
//	    Type:             manifest.TypeWidget,
//	    DefaultNamespace: "crucible",
//	})
package resource
