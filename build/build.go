package build

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/paths"
	"github.com/cruciblehq/crux/resource"
)

// Options for building a Crucible resource.
type Options struct {
	Manifest         string // Path to the manifest file.
	Output           string // Directory where built artifacts are placed.
	Registry         string // Hub registry URL (required for services to fetch runtimes).
	DefaultNamespace string // Default namespace for resource identifiers.
}

// Result of building a Crucible resource.
type Result struct {
	Output   string             // Path where the artifacts were written.
	Manifest *manifest.Manifest // The parsed manifest used for the build.
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

	context := filepath.Dir(opts.Manifest)

	var builder Builder

	switch man.Resource.Type {
	case resource.TypeRuntime:
		builder = NewRuntimeBuilder(opts.Registry, opts.DefaultNamespace, context)
	case resource.TypeService:
		builder = NewServiceBuilder(opts.Registry, opts.DefaultNamespace, context)
	case resource.TypeWidget:
		builder = NewWidgetBuilder()
	default:
		return nil, ErrInvalidResourceType
	}

	result, err := builder.Build(ctx, *man, opts.Output)
	if err != nil {
		return nil, err
	}

	result.Manifest = man
	return result, nil
}
