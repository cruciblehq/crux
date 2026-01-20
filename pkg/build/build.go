package build

import (
	"context"
	"os"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/paths"
	"github.com/cruciblehq/protocol/pkg/manifest"
)

const (

	// Directory where built artifacts are placed
	Dist = "dist"

	// The path of the manifest file within a Crucible resource project.
	Manifestfile = "crucible.yaml"
)

// Builds the Crucible resource located in the current working directory.
//
// It reads the manifest file, selects the appropriate builder based on the
// resource type, and invokes the build process. The built artifacts are placed
// in the [Dist] directory.
func Build(ctx context.Context) error {

	// Load manifest options
	man, err := manifest.Read(Manifestfile)
	if err != nil {
		return err
	}

	// Ensure output directory exists (same for all builders)
	if err := os.MkdirAll(Dist, paths.DefaultDirMode); err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}

	var builder Builder

	switch man.Resource.Type {
	case "widget":
		builder = NewWidgetBuilder()
	case "service":
		builder = NewServiceBuilder()
	default:
		return ErrInvalidResourceType
	}

	return builder.Build(ctx, *man)
}
