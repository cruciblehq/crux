package resource

import (
	"context"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/spec/manifest"
)

// Creates a new [Options] with the given defaults.
//
// Both defaultRegistry and defaultNamespace are required.
func NewOptions(client BuildClient, defaultRegistry, defaultNamespace string) Options {
	return Options{
		Client:           client,
		DefaultRegistry:  defaultRegistry,
		DefaultNamespace: defaultNamespace,
	}
}

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
	// falling back to the default registry configured in [Options].
	Push(ctx context.Context, m manifest.Manifest, packagePath string) error
}

// Configures a [Builder] obtained through [ResolveBuilder].
type Options struct {
	Client           BuildClient // Connection to the cruxd instance.
	DefaultRegistry  string      // Fallback registry for unqualified references.
	DefaultNamespace string      // Fallback namespace for unqualified references.
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

	defaults, err := NewDefaults(opts.DefaultRegistry, opts.DefaultNamespace)
	if err != nil {
		return nil, nil, err
	}

	workdir := filepath.Dir(manifestPath)

	var b Builder
	switch man.Resource.Type {
	case manifest.TypeRuntime:
		b = NewRuntimeBuilder(opts.Client, defaults, workdir)

	case manifest.TypeService:
		b = NewServiceBuilder(opts.Client, defaults, workdir)

	case manifest.TypeWidget:
		b = NewWidgetBuilder(defaults)

	case manifest.TypeMachine:
		b = NewMachineBuilder(opts.Client, defaults, workdir)

	default:
		return nil, nil, crex.Wrapf(ErrResolveBuilder, "resource type %q is not supported", man.Resource.Type)
	}

	return man, b, nil
}
