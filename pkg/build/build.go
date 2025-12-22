package build

import (
	"context"
	"errors"
	"os"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/manifest"
	"github.com/cruciblehq/crux/pkg/paths"
)

var (
	ErrBuildFailed         = errors.New("build failed")
	ErrInvalidResourceType = errors.New("invalid resource type")
	ErrInvalidPath         = errors.New("invalid path")
)

const (

	// Directory where built artifacts are placed
	Dist = "dist"
)

func Build(ctx context.Context) error {

	// Load manifest options
	man, err := manifest.Read()
	if err != nil {
		return err
	}

	// Ensure output directory exists (same for all builders)
	if err := os.MkdirAll(Dist, paths.DefaultDirMode); err != nil {
		return crex.Wrap(ErrBuildFailed, err)
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
