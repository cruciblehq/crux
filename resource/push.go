package resource

import (
	"context"

	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/registry"
)

// Pushes a packaged resource archive to the registry.
//
// The resource name, type, and version are taken from the manifest. The
// package is the .tar.zst archive produced by [Pack].
func Push(ctx context.Context, src registry.Source, m manifest.Manifest, packagePath string) error {
	return src.Push(ctx, m.Resource.Name, string(m.Resource.Type), m.Resource.Version, packagePath)
}
