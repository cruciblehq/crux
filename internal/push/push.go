package push

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/cache"
	"github.com/cruciblehq/crux/internal/registry"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/reference"
)

// Options for pushing a package to the Hub registry.
type PushOptions struct {
	Registry         string // Hub registry URL.
	Manifestfile     string // Path to the manifest file.
	Package          string // Path to the package archive.
	DefaultNamespace string // Default namespace for resource identifiers.
}

// Pushes a resource package to the Hub registry.
func Push(ctx context.Context, opts PushOptions) error {
	if err := validatePackage(opts.Package); err != nil {
		return err
	}

	man, err := loadManifest(opts.Manifestfile)
	if err != nil {
		return err
	}

	refOpts, err := reference.NewIdentifierOptions(opts.Registry, opts.DefaultNamespace)
	if err != nil {
		return err
	}

	id, err := reference.ParseIdentifier(man.Resource.Name, string(man.Resource.Type), refOpts)
	if err != nil {
		return err
	}

	client := registry.NewClient(opts.Registry, nil)

	if err := verifyNamespace(ctx, client, id.Namespace()); err != nil {
		return err
	}

	if err := ensureResource(ctx, client, id.Namespace(), id.Name(), man); err != nil {
		return err
	}

	if err := createVersion(ctx, client, id.Namespace(), id.Name(), man.Resource.Version); err != nil {
		return err
	}

	return uploadPackage(ctx, client, id.Namespace(), id.Name(), man.Resource.Version, opts.Package)
}

// Validates that the package file exists.
//
// The package file is expected to be at the default location. If not found, an
// error is returned prompting the user to package the resource first. The
// package's format and integrity are not validated.
func validatePackage(packageOutput string) error {
	f, err := os.Open(packageOutput)
	if err != nil {
		if os.IsNotExist(err) {
			return crex.UserError("package not found", fmt.Sprintf("%s does not exist", packageOutput)).
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
func loadManifest(manifestfile string) (*manifest.Manifest, error) {
	data, err := os.ReadFile(manifestfile)
	if err != nil {
		return nil, crex.UserError("failed to read manifest", err.Error()).
			Fallback("Ensure crucible.yaml exists and is valid.").
			Err()
	}

	man, err := manifest.Decode(data)
	if err != nil {
		return nil, crex.UserError("failed to read manifest", err.Error()).
			Fallback("Ensure crucible.yaml exists and is valid.").
			Err()
	}

	return man, nil
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
			Fallback("Create the namespace first.").
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

	var regErr *registry.Error
	if !errors.As(err, &regErr) || regErr.Code != registry.ErrorCodeNotFound {
		return crex.UserError("failed to check resource", err.Error()).
			Fallback("Crucible's Hub connectivity may be impaired. Try again later.").
			Err()
	}

	resInfo := registry.ResourceInfo{
		Name:        resource,
		Type:        string(man.Resource.Type),
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
// the registry. After successful upload, updates the local cache.
func uploadPackage(ctx context.Context, client *registry.Client, namespace, resource, version, packageOutput string) error {
	archive, err := os.Open(packageOutput)
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

	// Update local cache with pushed package
	if err := updateLocalCache(ctx, namespace, resource, version, packageOutput); err != nil {
		// Log warning but don't fail the push - the remote was updated successfully
		slog.Warn("failed to update local cache", "error", err)
	}

	return nil
}

// Adds the pushed package to the local cache.
//
// This ensures the local cache is in sync with the remote after a push,
// avoiding the need to re-download the package if it's needed locally.
func updateLocalCache(ctx context.Context, namespace, resource, version, packagePath string) error {
	localCache, err := cache.Open(ctx, nil)
	if err != nil {
		return err
	}
	defer localCache.Close()

	archive, err := os.Open(packagePath)
	if err != nil {
		return err
	}
	defer archive.Close()

	_, err = localCache.Put(ctx, namespace, resource, version, archive)
	return err
}
