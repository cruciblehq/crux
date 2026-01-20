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

// Options for pushing a package to the Hub registry.
type PushOptions struct {
	HubURL   string // Hub registry URL
	Resource string // Resource identifier (namespace/name)
}

// Pushes a resource package to the Hub registry.
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

// Validates that the package file exists.
//
// The package file is expected to be at the default location. If not found, an
// error is returned prompting the user to package the resource first. The
// package's format and integrity are not validated.
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

// Loads and validates the manifest.
//
// Ensures that crucible.yaml exists and is valid. If not, an error is returned.
// Returns the loaded manifest on success.
func loadManifest() (*manifest.Manifest, error) {
	man, err := manifest.Read(pack.Manifestfile)
	if err != nil {
		return nil, crex.UserError("failed to read manifest", err.Error()).
			Fallback("Ensure crucible.yaml exists and is valid.").
			Err()
	}
	return man, nil
}

// Parses namespace and resource name from the resource argument.
//
// The expected format is "namespace/name". If the format is invalid, an error
// is returned.
func parseResource(resource string) (namespace, name string, err error) {
	parts := strings.Split(resource, "/")
	if len(parts) != 2 {
		return "", "", crex.UserError("invalid resource format", "expected namespace/name").
			Fallback("Use the expected format namespace/name.").
			Err()
	}
	return parts[0], parts[1], nil
}

// Verifies that the namespace exists.
//
// Makes a call to the registry to check if the specified namespace exists. If
// the namespace does not exist, an error is returned prompting the user to
// create it first.
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
		Fallback("Crucible's Hub connectivity may be impaired. Try again later.").
		Err()
}

// Ensures the resource exists, creating it if necessary.
//
// Checks if the resource exists in the registry. If not, it creates the
// resource using the information from the manifest.
func ensureResource(ctx context.Context, client *registry.Client, namespace, resource string, man *manifest.Manifest) error {
	_, err := client.ReadResource(ctx, namespace, resource)
	if err == nil {
		return nil // Resource exists
	}

	// Check if error is "not found"
	var regErr *registry.Error
	if !errors.As(err, &regErr) || regErr.Code != registry.ErrorCodeNotFound {
		return crex.UserError("failed to check resource", err.Error()).
			Fallback("Crucible's Hub connectivity may be impaired. Try again later.").
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

// Creates a new version in the registry.
//
// Attempts to create the specified version for the resource. If the version
// already exists, an error is returned prompting the user to increment the
// version in the manifest.
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

// Uploads the package archive to the registry.
//
// Opens the package file and uploads it to the specified resource version in
// the registry.
func uploadPackage(ctx context.Context, client *registry.Client, namespace, resource, version string) error {
	archive, err := os.Open(pack.PackageOutput)
	if err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
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
