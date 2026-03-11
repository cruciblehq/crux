package resource

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/reference"
)

// [Builder] for Crucible machine images.
//
// Machine images are built externally and placed in the build directory before
// running crux build. The builder validates that the expected artifacts already
// exist, writes the resolved manifest, and delegates packing and pushing to the
// generic helpers. Future revisions should consider supporting an integrated
// build process for machine images.
type MachineBuilder struct {
	source Source // Source for push operations.
}

// Returns a [MachineBuilder].
func NewMachineBuilder(source Source) *MachineBuilder {
	return &MachineBuilder{
		source: source,
	}
}

// Validates pre-built machine images and writes the resolved manifest.
//
// Unlike other builders, Build does not invoke a build tool. It checks that
// the expected QCOW2 images already exist in the output directory and writes
// the crucible.yaml manifest file with resolved references.
func (mb *MachineBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	if _, ok := m.Config.(*manifest.Machine); !ok {
		return nil, crex.ProgrammingError("build failed", "an internal configuration type mismatch occurred").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	for _, img := range []string{manifest.MachineImageAarch64, manifest.MachineImageX86_64} {
		p := filepath.Join(output, img)
		if _, err := os.Stat(p); err != nil {
			return nil, crex.UserError("machine image not found", img+" does not exist in the build directory").
				Fallback("Place the pre-built machine images in the build directory before running 'crux build'.").
				Cause(err).
				Err()
		}
	}

	if _, err := reference.ParseIdentifier(m.Resource.Name, string(m.Resource.Type)); err != nil {
		return nil, crex.UserError("invalid resource name", "could not parse the resource identifier").
			Fallback("Check the resource name in crucible.yaml.").
			Cause(err).
			Err()
	}

	if err := WriteManifest(&m, output); err != nil {
		return nil, err
	}

	return &BuildResult{
		Output:   output,
		Manifest: &m,
	}, nil
}

// Validates that the build directory contains the expected machine artifacts.
//
// A valid machine build directory must contain crucible.yaml, aarch64.qcow2,
// and x86_64.qcow2.
func (mb *MachineBuilder) Validate(buildDir string) error {
	manifestPath := filepath.Join(buildDir, manifest.ManifestFile)
	if _, err := os.Stat(manifestPath); err != nil {
		return crex.UserError("manifest not found", "build/crucible.yaml does not exist").
			Fallback("Run 'crux build' first to generate the build artifacts.").
			Cause(err).
			Err()
	}

	for _, img := range []string{manifest.MachineImageAarch64, manifest.MachineImageX86_64} {
		p := filepath.Join(buildDir, img)
		if _, err := os.Stat(p); err != nil {
			return crex.UserError("machine image not found", "build/"+img+" does not exist").
				Fallback("Place the pre-built machine images in the build directory before running 'crux build'.").
				Cause(err).
				Err()
		}
	}

	return nil
}

// Packages the machine's build output into a distributable archive.
//
// The build directory must contain the QCOW2 images and crucible.yaml.
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
	return push(ctx, m, packagePath, mb.source)
}
