package build

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/protocol/pkg/manifest"
	"github.com/cruciblehq/protocol/pkg/oci"
)

const (

	// Relative path to the standardized service image location within dist/
	ServiceImagePath = "image.tar"
)

// Builder for Crucible services.
type ServiceBuilder struct{}

// Creates a new instance of [ServiceBuilder].
func NewServiceBuilder() *ServiceBuilder {
	return &ServiceBuilder{}
}

// Builds a Crucible service resource based on the provided manifest.
//
// Service resources reference pre-built container images. This method validates
// that the specified image exists and prepares it for packaging by copying it
// to the standardized dist/ output location (dist/image.tar).
func (sb *ServiceBuilder) Build(ctx context.Context, m manifest.Manifest) error {

	// Correct manifest type?
	service, ok := m.Config.(*manifest.Service)
	if !ok {
		return crex.ProgrammingError("an internal configuration type mismatch occurred", "unexpected manifest type").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	// Check if image path is specified
	if service.Build.Image == "" {
		return crex.UserError("service image not specified", "no image path in manifest").
			Fallback("Add a build image to your manifest.").
			Err()
	}

	// Validate it's a multi-platform OCI image
	imagePath := service.Build.Image
	if err := validateOCIMultiPlatform(imagePath); err != nil {
		if os.IsNotExist(err) {
			return crex.UserError("service image not found", "image does not exist at specified path").
				Cause(err).
				Fallback("Either the image path is incorrect or the image has not been built yet. Try building your service image first and make sure the image file exists at the specified path.").
				Err()
		}
		return err
	}

	// Copy to standardized output location (always dist/image.tar)
	destPath := filepath.Join(Dist, ServiceImagePath)
	if err := copyFile(imagePath, destPath); err != nil {
		return crex.Wrap(ErrBuildFailed, err)
	}

	return nil
}

// Copies a file from src to dst.
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// Validates that an OCI tarball contains required platforms.
func validateOCIMultiPlatform(path string) error {
	err := oci.ValidateMultiPlatform(path)
	if err == nil {
		return nil
	}

	if err == oci.ErrSinglePlatform {
		return crex.UserError("incomplete platform support", err.Error()).
			Fallback(fmt.Sprintf("Build a multi-platform image with required platforms %v", oci.RequiredPlatforms())).
			Err()
	}

	if err == oci.ErrInvalidImage {
		return crex.UserError("invalid OCI image", err.Error()).
			Fallback("The image file does not appear to be a valid OCI image. Make sure you're exporting with type=oci.").
			Err()
	}

	return crex.Wrap(ErrBuildFailed, err)
}
