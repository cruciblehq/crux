package build

import (
	"context"
	"os"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/manifest"
	"github.com/cruciblehq/crux/pkg/paths"
)

const (

	// Directory where built artifacts are placed
	Dist = "dist"
)

// Builds the provided manifest using the appropriate builder based on
// the resource type.
//
// It selects the appropriate builder and delegates the build process.
func Build(ctx context.Context, m manifest.Manifest) error {

	// Ensure output directory exists (same for all builders)
	if err := os.MkdirAll(Dist, paths.DefaultDirMode); err != nil {
		return crex.ProgrammingError("build failed", "failed to create output directory").
			Cause(err).
			Err()
	}

	var builder Builder

	switch m.Resource.Type {
	case "widget":
		builder = NewWidgetBuilder()
	case "service":
		builder = NewServiceBuilder()
	default:
		return crex.UserErrorf("build failed", "invalid resource type '%s'", m.Resource.Type).
			Fallback("Change your manifest to use a supported resource type.").
			Err()
	}

	return builder.Build(ctx, m)
}
