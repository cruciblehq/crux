package resource

import (
	"context"
	"io"
	"path/filepath"

	"github.com/cruciblehq/crux/resource/blueprint"
	"github.com/cruciblehq/crux/resource/runtime"
	"github.com/cruciblehq/crux/resource/service"
	"github.com/cruciblehq/crux/resource/widget"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/registry"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
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
// The caller creates the output directory.
func Build(ctx context.Context, m manifest.Manifest, src registry.Source, workdir, env, output string) (*BuildResult, error) {
	switch m.Resource.Type {
	case manifest.TypeRuntime:
		return buildRuntime(ctx, &m, src, workdir, output)
	case manifest.TypeService:
		return buildService(ctx, &m, src, workdir, output)
	case manifest.TypeWidget:
		return buildWidget(ctx, &m, output)
	case manifest.TypeBlueprint:
		return buildBlueprint(ctx, &m, src, env, output)
	default:
		return nil, crex.Newf(ErrUnsupportedType, "resource type %q is not supported", m.Resource.Type)
	}
}

// Builds a runtime resource and writes its image and manifest to output.
func buildRuntime(ctx context.Context, m *manifest.Manifest, src registry.Source, workdir, output string) (*BuildResult, error) {
	cfg, err := manifest.As[*manifest.Runtime](m)
	if err != nil {
		return nil, err
	}
	err = writeArtifact(output, manifest.ImageFile, func(w io.Writer) error {
		return runtime.NewBuilder(src, workdir).Build(ctx, cfg, w)
	})
	if err != nil {
		return nil, err
	}
	return writeResult(m, output)
}

// Builds a service resource and writes its image and manifest to output.
func buildService(ctx context.Context, m *manifest.Manifest, src registry.Source, workdir, output string) (*BuildResult, error) {
	cfg, err := manifest.As[*manifest.Service](m)
	if err != nil {
		return nil, err
	}
	err = writeArtifact(output, manifest.ImageFile, func(w io.Writer) error {
		return service.NewBuilder(src, workdir).Build(ctx, cfg, w)
	})
	if err != nil {
		return nil, err
	}
	return writeResult(m, output)
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

// Builds a blueprint resource and writes its plan and manifest to output.
func buildBlueprint(ctx context.Context, m *manifest.Manifest, src registry.Source, env, output string) (*BuildResult, error) {
	cfg, err := manifest.As[*manifest.Blueprint](m)
	if err != nil {
		return nil, err
	}
	err = writeArtifact(output, manifest.PlanFile, func(w io.Writer) error {
		return blueprint.NewBuilder(src, env, manifest.Compute{}).Build(ctx, cfg, w)
	})
	if err != nil {
		return nil, err
	}
	return writeResult(m, output)
}

// Writes the file name in dir with what write produces.
//
// The file appears only when write succeeds, and then replaces any earlier
// one. Errors from write are returned as they are.
func writeArtifact(dir, name string, write func(io.Writer) error) error {
	a, err := file.CreateAtomic(dir, file.DefaultFileMode)
	if err != nil {
		return outputError(dir, err)
	}
	defer a.Abort()
	if err := write(a); err != nil {
		return err
	}
	if err := a.Commit(filepath.Join(dir, name)); err != nil {
		return outputError(dir, err)
	}
	return nil
}

// Writes the resolved manifest to dir and returns the build result.
func writeResult(m *manifest.Manifest, dir string) (*BuildResult, error) {
	if err := WriteManifest(dir, m); err != nil {
		return nil, outputError(dir, err)
	}
	return &BuildResult{Output: dir, Manifest: m}, nil
}

// Returns the error for a failed write of build output to dir.
func outputError(dir string, err error) error {
	return crex.SystemError("cannot write build output", "failed to write to the build directory").
		Recoveryf("Make sure you have write access to %s, then try again.", dir).
		Cause(crex.Wrap(ErrFileSystemOperation, err)).
		Err()
}
