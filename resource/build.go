package resource

import (
	"context"

	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/crux/resource/affordance"
	"github.com/cruciblehq/crux/resource/blueprint"
	"github.com/cruciblehq/crux/resource/runtime"
	"github.com/cruciblehq/crux/resource/service"
	"github.com/cruciblehq/crux/resource/widget"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/crex"
)

// Holds the output of a successful [Build] call.
type BuildResult struct {
	Output   string             // Directory where the build artifacts were written.
	Manifest *manifest.Manifest // The fully resolved manifest used for the build.
}

// Compiles a resource according to its manifest.
//
// Dispatches to the matching type builder based on the manifest type, then
// writes the resolved manifest to the output directory alongside any type
// specific artifacts. workdir is the manifest directory; the local source
// paths in copy steps are resolved relative to it for runtime and service
// builds. env selects the blueprint environment and is ignored by other types.
// The output directory is created by the caller and guaranteed to be empty.
func Build(ctx context.Context, m manifest.Manifest, src hub.Source, workdir, env, output string) (*BuildResult, error) {
	switch m.Resource.Type {
	case manifest.TypeRuntime:
		return buildRuntime(ctx, &m, src, workdir, output)
	case manifest.TypeService:
		return buildService(ctx, &m, src, workdir, output)
	case manifest.TypeWidget:
		return buildWidget(ctx, &m, output)
	case manifest.TypeAffordance:
		return buildAffordance(ctx, &m, src, output)
	case manifest.TypeBlueprint:
		return buildBlueprint(ctx, &m, src, env, output)
	default:
		return nil, crex.Newf(ErrUnsupportedType, "resource type %q is not supported", m.Resource.Type)
	}
}

// Builds a runtime resource and writes the resolved manifest to the build dir.
func buildRuntime(ctx context.Context, m *manifest.Manifest, src hub.Source, workdir, output string) (*BuildResult, error) {
	cfg, err := manifest.As[*manifest.Runtime](m)
	if err != nil {
		return nil, err
	}
	buildDir, err := runtime.NewBuilder(src, workdir).Build(ctx, cfg, output)
	if err != nil {
		return nil, err
	}
	return writeResult(m, buildDir)
}

// Builds a service resource and writes the resolved manifest to the build dir.
func buildService(ctx context.Context, m *manifest.Manifest, src hub.Source, workdir, output string) (*BuildResult, error) {
	cfg, err := manifest.As[*manifest.Service](m)
	if err != nil {
		return nil, err
	}
	buildDir, err := service.NewBuilder(src, workdir).Build(ctx, cfg, output)
	if err != nil {
		return nil, err
	}
	return writeResult(m, buildDir)
}

// Builds a widget resource and writes the resolved manifest to output.
func buildWidget(ctx context.Context, m *manifest.Manifest, output string) (*BuildResult, error) {
	cfg, err := manifest.As[*manifest.Widget](m)
	if err != nil {
		return nil, err
	}
	if _, err := widget.NewBuilder().Build(ctx, cfg, output); err != nil {
		return nil, err
	}
	return writeResult(m, output)
}

// Builds an affordance resource, compiling every grant in every scope, then
// writes the resolved manifest to output.
func buildAffordance(ctx context.Context, m *manifest.Manifest, src hub.Source, output string) (*BuildResult, error) {
	cfg, err := manifest.As[*manifest.Affordance](m)
	if err != nil {
		return nil, err
	}
	b := affordance.NewBuilder()
	for _, scope := range cfg.Scopes {
		for _, g := range scope.Grants {
			if err := b.Build(ctx, g, src); err != nil {
				return nil, err
			}
		}
	}
	m.Config = &manifest.Affordance{Schema: cfg.Schema, Scopes: cfg.Scopes}
	return writeResult(m, output)
}

// Builds a blueprint resource and writes the resolved manifest to output.
func buildBlueprint(ctx context.Context, m *manifest.Manifest, src hub.Source, env, output string) (*BuildResult, error) {
	cfg, err := manifest.As[*manifest.Blueprint](m)
	if err != nil {
		return nil, err
	}
	if err := blueprint.NewBuilder(src, env).Build(ctx, cfg, output); err != nil {
		return nil, err
	}
	return writeResult(m, output)
}

// Writes the resolved manifest to dir and returns the build result.
func writeResult(m *manifest.Manifest, dir string) (*BuildResult, error) {
	if err := manifest.WriteAt(m, dir); err != nil {
		return nil, crex.SystemError("cannot write build output", "failed to write the resolved manifest to the build directory").
			Recoveryf("Make sure you have write access to %s, then try again.", dir).
			Cause(err).
			Err()
	}
	return &BuildResult{Output: dir, Manifest: m}, nil
}
