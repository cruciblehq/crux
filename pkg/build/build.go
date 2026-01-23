package build

import (
	"context"
	"os"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/paths"
	"github.com/cruciblehq/protocol/pkg/manifest"
	"github.com/cruciblehq/protocol/pkg/resource"
)

// Options for building a Crucible resource.
type Options struct {
	Manifest string // Path to the manifest file.
	Output   string // Directory where built artifacts are placed.
}

// Result of building a Crucible resource.
type Result struct {
	Output string // Path where the artifacts were written.
}

// Builds the Crucible resource.
//
// It reads the manifest file, selects the appropriate builder based on the
// resource type, and invokes the build process. The built artifacts are placed
// in the directory specified by opts.OutputPath.
func Build(ctx context.Context, opts Options) (*Result, error) {

	// Load manifest options
	man, err := manifest.Read(opts.Manifest)
	if err != nil {
		return nil, err
	}

	// Ensure output directory exists (same for all builders)
	if err := os.MkdirAll(opts.Output, paths.DefaultDirMode); err != nil {
		return nil, crex.Wrap(ErrFileSystemOperation, err)
	}

	var builder Builder

	switch resource.Type(man.Resource.Type) {
	case resource.TypeWidget:
		builder = NewWidgetBuilder()
	case resource.TypeService:
		builder = NewServiceBuilder()
	default:
		return nil, ErrInvalidResourceType
	}

	result, err := builder.Build(ctx, *man, opts.Output)
	if err != nil {
		return nil, err
	}

	return result, nil
}
