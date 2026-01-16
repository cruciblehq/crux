package push

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/pack"
	"github.com/cruciblehq/protocol/pkg/manifest"
	"github.com/cruciblehq/protocol/pkg/registry"
)

var (
	ErrPushFailed      = errors.New("push failed")
	ErrInvalidResource = errors.New("invalid resource format")
)

// Options for pushing a package to the Hub registry.
type PushOptions struct {
	HubURL   string // Hub registry URL
	Resource string // Resource identifier (namespace/name)
}

// Push pushes a resource package to the Hub registry.
func Push(ctx context.Context, opts PushOptions) error {
	// Validate package exists
	if err := validatePackage(); err != nil {
		return err
	}

	// Load manifest
	man, err := loadManifest()
	if err != nil {
		return err
	}

	// Parse resource identifier
	namespace, resourceName, err := parseResource(opts.Resource)
	if err != nil {
		return err
	}

	// Create registry client
	client := registry.NewClient(opts.HubURL, nil)

	// Verify namespace exists
	if err := verifyNamespace(ctx, client, namespace); err != nil {
		return err
	}

	// Ensure resource exists (create if not)
	if err := ensureResource(ctx, client, namespace, resourceName, man); err != nil {
		return err
	}

	// Create version
	if err := createVersion(ctx, client, namespace, resourceName, man.Resource.Version); err != nil {
		return err
	}

	// Upload package
	return uploadPackage(ctx, client, namespace, resourceName, man.Resource.Version)
}

// validatePackage validates that the package file exists.
func validatePackage() error {
	f, err := os.Open(pack.PackageOutput)
	if err != nil {
		if os.IsNotExist(err) {
			return crex.UserError("package not found", fmt.Sprintf("%s does not exist", pack.PackageOutput)).
				Fallback("Run 'crux pack' first to create the package.").
				Err()
		}
		return err
	}
	f.Close()
	return nil
}

// loadManifest loads and validates the manifest.
func loadManifest() (*manifest.Manifest, error) {
	man, err := manifest.Read(pack.Manifestfile)
	if err != nil {
		return nil, crex.UserError("failed to read manifest", err.Error()).
			Fallback("Ensure crucible.yaml exists and is valid.").
			Err()
	}
	return man, nil
}

// parseResource parses namespace and resource name from the resource argument.
func parseResource(resource string) (namespace, resourceName string, err error) {
	parts := strings.Split(resource, "/")
	if len(parts) != 2 {
		return "", "", crex.UserError("invalid resource format", "expected namespace/name").
			Fallback("Use format: crux push namespace/name").
			Err()
	}
	return parts[0], parts[1], nil
}

// verifyNamespace verifies that the namespace exists.
func verifyNamespace(ctx context.Context, client *registry.Client, namespace string) error {
	_, err := client.ReadNamespace(ctx, namespace)
	if err == nil {
		return nil // Namespace exists
	}

	var regErr *registry.Error
	if errors.As(err, &regErr) && regErr.Code == registry.ErrorCodeNotFound {
		return crex.UserError("namespace not found", fmt.Sprintf("namespace '%s' does not exist", namespace)).
			Fallback("Create the namespace first using the Hub administration tools.").
			Err()
	}

	return crex.UserError("failed to check namespace", err.Error()).
		Fallback("Check Hub connectivity.").
		Err()
}

// ensureResource ensures the resource exists, creating it if necessary.
func ensureResource(ctx context.Context, client *registry.Client, namespace, resource string, man *manifest.Manifest) error {
	_, err := client.ReadResource(ctx, namespace, resource)
	if err == nil {
		return nil // Resource exists
	}

	// Check if error is "not found"
	var regErr *registry.Error
	if !errors.As(err, &regErr) || regErr.Code != registry.ErrorCodeNotFound {
		return crex.UserError("failed to check resource", err.Error()).
			Fallback("Check Hub connectivity.").
			Err()
	}

	// Create resource
	resInfo := registry.ResourceInfo{
		Name:        resource,
		Type:        man.Resource.Type,
		Description: "",
	}
	_, err = client.CreateResource(ctx, namespace, resInfo)
	if err != nil {
		return crex.UserError("failed to create resource", err.Error()).
			Fallback("Check Hub connectivity and permissions.").
			Err()
	}

	return nil
}

// createVersion creates a new version in the registry.
func createVersion(ctx context.Context, client *registry.Client, namespace, resource, version string) error {
	versionInfo := registry.VersionInfo{
		String: version,
	}

	_, err := client.CreateVersion(ctx, namespace, resource, versionInfo)
	if err != nil {
		var regErr *registry.Error
		if errors.As(err, &regErr) && regErr.Code == registry.ErrorCodeVersionExists {
			return crex.UserError("version already exists", fmt.Sprintf("version %s already exists", version)).
				Fallback("Increment the version in crucible.yaml and rebuild.").
				Err()
		}
		return crex.UserError("failed to create version", err.Error()).
			Fallback("Check Hub connectivity and permissions.").
			Err()
	}

	return nil
}

// uploadPackage uploads the package archive to the registry.
func uploadPackage(ctx context.Context, client *registry.Client, namespace, resource, version string) error {
	archive, err := os.Open(pack.PackageOutput)
	if err != nil {
		return crex.Wrap(ErrPushFailed, err)
	}
	defer archive.Close()

	_, err = client.UploadArchive(ctx, namespace, resource, version, archive)
	if err != nil {
		return crex.UserError("failed to upload archive", err.Error()).
			Fallback("Check Hub connectivity and package integrity.").
			Err()
	}

	return nil
}
