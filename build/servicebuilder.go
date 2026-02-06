package build

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/pack"
	"github.com/cruciblehq/crux/manifest"
)

// Builder for Crucible services.
type ServiceBuilder struct{}

// Creates a new instance of [ServiceBuilder].
func NewServiceBuilder(registry string) *ServiceBuilder {
	return &ServiceBuilder{}
}

// Builds a Crucible service resource based on the provided manifest.
//
// Service resources require a pre-built image.tar file in the current directory.
// This method validates the image exists and copies it to the build/ directory.
func (sb *ServiceBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*Result, error) {
	service, ok := m.Config.(*manifest.Service)
	if !ok {
		return nil, crex.ProgrammingError("an internal configuration type mismatch occurred", "unexpected manifest type").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	if err := sb.validateManifest(service); err != nil {
		return nil, err
	}

	sourceImage := pack.ImageFile
	if _, err := os.Stat(sourceImage); os.IsNotExist(err) {
		return nil, crex.UserError("image file not found", "image.tar does not exist in the current directory").
			Fallback("Build the image first using Docker or another tool, then run crux build.").
			Err()
	}

	destImage := filepath.Join(output, pack.ImageFile)
	if err := copyFile(sourceImage, destImage); err != nil {
		return nil, crex.Wrap(ErrFileSystemOperation, err)
	}

	return &Result{Output: output}, nil
}

// Validates required fields in the service manifest.
func (sb *ServiceBuilder) validateManifest(_ *manifest.Service) error {
	return nil
}

// Copies a file from src to dst.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
