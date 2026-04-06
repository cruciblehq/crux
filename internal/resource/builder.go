package resource

import (
	"context"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
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

	// Verifies that a build directory contains the expected artifacts for
	// the resource type.
	//
	// The build directory must contain a crucible.yaml whose resource type
	// matches the builder. Each resource type then checks for its own
	// type-specific artifacts (e.g. image.tar for services, index.js for
	// widgets, plan.yaml for blueprints).
	Verify(buildDir string) error

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
	DefaultRegistry  string // Fallback registry for unqualified references.
	DefaultNamespace string // Fallback namespace for unqualified references.
}

// Creates a new [Options] with the given defaults.
//
// Both defaultRegistry and defaultNamespace are required.
func NewOptions(defaultRegistry, defaultNamespace string) (Options, error) {
	if defaultRegistry == "" {
		return Options{}, crex.Wrap(ErrMissingOption, ErrMissingRegistry)
	}
	if defaultNamespace == "" {
		return Options{}, crex.Wrap(ErrMissingOption, ErrMissingNamespace)
	}
	return Options{
		DefaultRegistry:  defaultRegistry,
		DefaultNamespace: defaultNamespace,
	}, nil
}

// Reads the manifest at the given path and returns the appropriate [Builder]
// for the resource type declared in it.
func ResolveBuilder(ctx context.Context, manifestPath string, opts Options) (*manifest.Manifest, Builder, error) {
	man, err := ReadManifest(manifestPath)
	if err != nil {
		return nil, nil, err
	}

	source, err := NewSource(opts.DefaultRegistry, opts.DefaultNamespace)
	if err != nil {
		return nil, nil, err
	}

	workdir := filepath.Dir(manifestPath)

	var b Builder
	switch man.Resource.Type {
	case manifest.TypeRuntime:
		b = NewRuntimeBuilder(source, workdir)

	case manifest.TypeService:
		b = NewServiceBuilder(source, workdir)

	case manifest.TypeWidget:
		b = NewWidgetBuilder(source)

	case manifest.TypeAffordance:
		b = NewAffordanceBuilder(source)

	case manifest.TypeBlueprint:
		b = NewBlueprintBuilder(source, "")

	default:
		return nil, nil, crex.Wrapf(ErrResolveBuilder, "resource type %q is not supported", man.Resource.Type)
	}

	return man, b, nil
}

// Extracts the typed config from a manifest.
//
// Returns the config cast to the expected type T, or a programming error if
// the manifest config does not match.
func manifestConfig[T any](m *manifest.Manifest) (T, error) {
	cfg, ok := m.Config.(T)
	if !ok {
		var zero T
		return zero, crex.ProgrammingError("build failed", "an internal configuration type mismatch occurred").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}
	return cfg, nil
}
