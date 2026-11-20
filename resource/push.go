package resource

import (
	"context"
	"io"

	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/spec/manifest"
)

// Pushes a packaged resource archive to the registry.
//
// The resource name, type, and version are taken from the manifest. archive is
// the .tar.zst archive produced by [Pack].
func Push(ctx context.Context, src hub.Source, m manifest.Manifest, archive io.ReadSeeker) error {
	return src.Push(ctx, m.Resource.Name, string(m.Resource.Type), m.Resource.Version, archive)
}
