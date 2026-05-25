package resource

import (
	"context"

	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/source"
)

// Handles artifact operations for a single resource type.
//
// The Handler interface defines the operations for managing resource artifacts.
// The specific logic for each operation is implemented by the resource-specific
// handlers (e.g. [ServiceHandler], [WidgetHandler], etc) which are resolved at
// runtime based on the resource type in the manifest.
type Handler interface {

	// Compiles the resource according to its manifest.
	//
	// Resolves all references (including the resource name) and writes the
	// manifest to the output directory alongside other build artifacts. Each
	// resource type may include additional type-specific artifacts and build
	// logic. The output parameter is a directory path where the handler should
	// write the build artifacts, including the manifest. The output directory
	// is created by the caller and guaranteed to be empty. BuildResult includes
	// the resolved manifest and the output directory path for use in subsequent
	// operations (e.g. Pack).
	Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error)

	// Verifies that a build directory contains the expected artifacts.
	//
	// The build directory must contain a manifest whose resource type matches
	// the handler. Each resource type then checks for its own type-specific
	// artifacts (e.g. image.tar for services, plan.yaml for blueprints, etc).
	Verify(buildDir string) error

	// Packages built artifacts into a distributable archive.
	//
	// The archive layout is type-specific: each resource type decides which
	// files are included and how they are structured. The resolved manifest
	// is included in the archive. The output extension must be .tar.zst.
	Pack(ctx context.Context, buildDir, output string) (*PackResult, error)

	// Pushes a packaged resource archive to the registry.
	//
	// The handler uses the source (passed at construction) to push the package
	// to the registry. The resource name, type, and version are taken from the
	// manifest. The package is the .tar.zst archive produced by Pack.
	Push(ctx context.Context, m manifest.Manifest, packagePath string) error
}

// Holds the output of a successful [Handler.Build] call.
type BuildResult struct {
	Output   string             // Directory where the build artifacts were written.
	Manifest *manifest.Manifest // The fully resolved manifest used for the build.
}

// Selects the appropriate handler for the given resource type.
//
// Not all parameters are used by every handler. workdir is only relevant for
// handlers that resolve paths relative to the source tree (runtime, service).
// env is only relevant for blueprints. src is used by all handlers that
// interact with the registry.
func ResolveHandler(resourceType manifest.ResourceType, src source.Source, workdir, env string) (Handler, error) {
	switch resourceType {
	case manifest.TypeRuntime:
		return NewRuntimeHandler(src, workdir), nil

	case manifest.TypeService:
		return NewServiceHandler(src, workdir), nil

	case manifest.TypeWidget:
		return NewWidgetHandler(src), nil

	case manifest.TypeAffordance:
		return NewAffordanceHandler(src), nil

	case manifest.TypeBlueprint:
		return NewBlueprintHandler(src, env), nil

	default:
		return nil, crex.Wrapf(ErrResolveHandler, "resource type %q is not supported", resourceType)
	}
}
