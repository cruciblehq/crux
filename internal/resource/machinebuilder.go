package resource

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/spec/manifest"
)

// Builder for Crucible machine resources.
//
// Machine resources define bootable VM images. Only the Build, Pack, and Push
// operations are supported. Lifecycle operations (Start, Stop, Exec, etc.)
// are not applicable because machines are provisioned through the provider
// layer, not managed as running resources.
type MachineBuilder struct {
	recipeBuilder
}

// Returns a [MachineBuilder] wired to the given daemon client.
//
// workdir is the directory containing the manifest and is used as the root
// for resolving copy sources during builds.
func NewMachineBuilder(client *daemon.Client, registry, defaultNamespace, workdir string) *MachineBuilder {
	return &MachineBuilder{
		recipeBuilder: recipeBuilder{
			client:           client,
			registry:         registry,
			defaultNamespace: defaultNamespace,
			workdir:          workdir,
		},
	}
}

// Builds a Crucible machine resource based on the provided manifest.
//
// The machine configuration is extracted and the shared recipe pipeline
// handles the build process. The built artifacts are placed in the directory
// specified by the output parameter.
func (mb *MachineBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	cfg, ok := m.Config.(*manifest.Machine)
	if !ok {
		return nil, crex.ProgrammingError("build failed", "an internal configuration type mismatch occurred").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	if _, err := m.ResolveName(mb.registry, mb.defaultNamespace); err != nil {
		return nil, crex.UserError("invalid resource name", "could not resolve the resource identifier").
			Fallback("Check the resource name in crucible.yaml.").
			Cause(err).
			Err()
	}

	result, err := mb.build(ctx, m, &cfg.Recipe, output, nil)
	if err != nil {
		return nil, err
	}

	if err := WriteManifest(&m, result.Output); err != nil {
		return nil, err
	}

	return result, nil
}

// Validates that the build directory contains the expected machine artifacts.
//
// A valid machine build directory must contain image.tar.
func (mb *MachineBuilder) Validate(buildDir string) error {
	manifestPath := filepath.Join(buildDir, manifest.ManifestFile)
	if _, err := os.Stat(manifestPath); err != nil {
		return crex.UserError("manifest not found", "build/crucible.yaml does not exist").
			Fallback("Run 'crux build' first to generate the build artifacts.").
			Cause(err).
			Err()
	}

	imagePath := filepath.Join(buildDir, manifest.ImageFile)
	if _, err := os.Stat(imagePath); err != nil {
		return crex.UserError("machine build output not found", "build/image.tar does not exist").
			Fallback("Run 'crux build' to prepare the machine image.").
			Cause(err).
			Err()
	}

	return nil
}

// Packages the machine's build output into a distributable archive.
//
// The build directory must contain image.tar.
func (mb *MachineBuilder) Pack(ctx context.Context, buildDir, output string) (*PackResult, error) {
	if err := mb.Validate(buildDir); err != nil {
		return nil, err
	}
	return pack(ctx, buildDir, output)
}

// Uploads a machine package archive to the Hub registry.
//
// packagePath must point to an archive created by [MachineBuilder.Pack].
func (mb *MachineBuilder) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return push(ctx, m, packagePath, mb.registry, mb.defaultNamespace)
}
