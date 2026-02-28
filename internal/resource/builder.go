package resource

import (
	"context"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/provider"
	"github.com/cruciblehq/spec/manifest"
)

// Handles artifact operations for a single Crucible resource type.
//
// Each resource type provides its own Builder implementation. A Builder is
// obtained via [ResolveBuilder]. Lifecycle operations (start, stop, exec,
// etc.) are handled by the provider layer, not by the Builder.
type Builder interface {

	// Compiles the resource according to its manifest and writes the resulting
	// artifacts to the output directory.
	//
	// Build resolves all references (including the resource name) using the
	// configured defaults and writes the fully resolved manifest as
	// crucible.yaml in the output directory alongside the build artifacts.
	Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error)

	// Validates that a build directory contains the expected artifacts for
	// the resource type.
	//
	// Each resource type defines what constitutes a valid build output. For
	// example, image-based resources (runtimes, services, machines) require
	// image.tar, while widgets require index.js.
	Validate(buildDir string) error

	// Packages built artifacts into a distributable archive.
	//
	// The archive layout is type-specific: each resource type decides which
	// files are included and how they are structured. The resolved manifest
	// written by [Builder.Build] is included in the archive as crucible.yaml.
	// The output extension must be .tar.zst.
	Pack(ctx context.Context, buildDir, output string) (*PackResult, error)

	// Pushes a packaged resource archive to the registry.
	//
	// packagePath must point to an archive created by [Builder.Pack]. The
	// target registry is determined by the resource name in the manifest,
	// falling back to the DefaultRegistry provided in [Options].
	Push(ctx context.Context, m manifest.Manifest, packagePath string) error
}

// Configures a [Builder] obtained through [ResolveBuilder].
//
// Compute and NodeID are required for resource types that interact with
// the daemon (services, runtimes, and machines). They can be zero-valued
// for widgets. DefaultRegistry and DefaultNamespace are required for Build
// and Push. Build uses them to resolve references and writes the fully
// resolved manifest to the build output directory. Push uses them to
// determine the target registry. Pack does not need them because it reads
// the already-resolved manifest from the build directory.
type Options struct {
	Compute          provider.ComputeService // Compute backend that hosts the cruxd daemon.
	NodeID           string                  // Compute node running cruxd.
	DefaultRegistry  string                  // Fallback registry URL.
	DefaultNamespace string                  // Fallback resource namespace.
}

// Holds the output of a successful [Builder.Build] call.
type BuildResult struct {
	Output   string             // Directory where the build artifacts were written.
	Manifest *manifest.Manifest // The fully resolved manifest used for the build.
}

// Reads the manifest at the given path and returns the appropriate [Builder]
// for the resource type declared in it.
func ResolveBuilder(ctx context.Context, manifestPath string, opts Options) (*manifest.Manifest, Builder, error) {
	man, err := ReadManifest(manifestPath)
	if err != nil {
		return nil, nil, err
	}

	workdir := filepath.Dir(manifestPath)

	var b Builder
	switch man.Resource.Type {
	case manifest.TypeRuntime:
		client, err := opts.Compute.Client(ctx, opts.NodeID)
		if err != nil {
			return nil, nil, crex.Wrap(ErrResolveBuilder, err)
		}
		b = NewRuntimeBuilder(client, opts.DefaultRegistry, opts.DefaultNamespace, workdir)

	case manifest.TypeService:
		client, err := opts.Compute.Client(ctx, opts.NodeID)
		if err != nil {
			return nil, nil, crex.Wrap(ErrResolveBuilder, err)
		}
		b = NewServiceBuilder(client, opts.DefaultRegistry, opts.DefaultNamespace, workdir)

	case manifest.TypeWidget:
		b = NewWidgetBuilder(opts.DefaultRegistry, opts.DefaultNamespace)

	case manifest.TypeMachine:
		client, err := opts.Compute.Client(ctx, opts.NodeID)
		if err != nil {
			return nil, nil, crex.Wrap(ErrResolveBuilder, err)
		}
		b = NewMachineBuilder(client, opts.DefaultRegistry, opts.DefaultNamespace, workdir)

	default:
		return nil, nil, crex.Wrapf(ErrResolveBuilder, "resource type %q is not supported", man.Resource.Type)
	}

	return man, b, nil
}
