package build

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/spec/manifest"
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

	data, err := os.ReadFile(opts.Manifest)
	if err != nil {
		return nil, crex.Wrap(ErrBuild, err)
	}

	man, err := manifest.Decode(data)
	if err != nil {
		return nil, crex.Wrap(ErrBuild, err)
	}

	if err := os.MkdirAll(opts.Output, paths.DefaultDirMode); err != nil {
		return nil, crex.Wrap(ErrFileSystemOperation, err)
	}

	context := filepath.Dir(opts.Manifest)

	var builder Builder

	switch man.Resource.Type {
	case manifest.TypeRuntime:
		client := daemon.NewClient()
		builder = NewRuntimeBuilder(client, opts.Registry, opts.DefaultNamespace, context)
	case manifest.TypeService:
		client := daemon.NewClient()
		builder = NewServiceBuilder(client, opts.Registry, opts.DefaultNamespace, context)
	case manifest.TypeWidget:
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
