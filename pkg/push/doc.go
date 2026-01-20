// Package push provides functionality for publishing Crucible resources to
// a Hub registry.
//
// The package handles the complete workflow of uploading a packaged resource
// to a Crucible Hub, including namespace verification, resource creation,
// version management, and package upload. It expects a package.tar.zst file
// created by the pack command.
//
// Resources are identified using the format "namespace/name". The namespace
// must exist in the Hub before pushing a resource.
//
// Example usage:
//
//	opts := push.PushOptions{
//	    HubURL:   "http://hub.cruciblehq.xyz:8080",
//	    Resource: "myorg/mywidget",
//	}
//	if err := push.Push(ctx, opts); err != nil {
//	    return err
//	}
//
// The package is published to the Hub and becomes available for deployment
// and sharing with other Crucible users.
package push
