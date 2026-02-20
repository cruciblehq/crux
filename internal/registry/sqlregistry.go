package registry

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"os"
	"sync"
)

const (

	// Namespace operation error messages
	errMsgCreateNamespace       = "unable to create namespace due to internal error"
	errMsgRetrieveNamespace     = "unable to retrieve namespace information"
	errMsgSaveNamespaceChanges  = "unable to save namespace changes"
	errMsgDeleteNamespace       = "unable to delete namespace - it may contain resources"
	errMsgRetrieveNamespaceList = "unable to retrieve namespace list"
	errMsgNamespaceNotFound     = "namespace not found"
	errMsgNamespaceExists       = "namespace already exists"

	// Resource operation error messages
	errMsgCreateResource       = "unable to create resource due to internal error"
	errMsgRetrieveResource     = "unable to retrieve resource information"
	errMsgSaveResourceChanges  = "unable to save resource changes"
	errMsgDeleteResource       = "unable to delete resource - it may contain versions or channels"
	errMsgRetrieveResourceList = "unable to retrieve resource list for namespace"
	errMsgResourceNotFound     = "resource not found"
	errMsgResourceExists       = "resource already exists"

	// Version operation error messages
	errMsgCreateVersion       = "unable to create version due to internal error"
	errMsgRetrieveVersion     = "unable to retrieve version information"
	errMsgSaveVersionChanges  = "unable to save version changes"
	errMsgDeleteVersion       = "unable to delete version - it may be referenced by channels"
	errMsgRetrieveVersionList = "unable to retrieve version list for resource"
	errMsgVersionNotFound     = "version not found"
	errMsgVersionExists       = "version already exists"

	// Archive operation error messages
	errMsgStoreArchive      = "unable to store archive metadata"
	errMsgUpdateArchive     = "unable to update archive metadata"
	errMsgAccessArchiveFile = "unable to access archive file - file may be missing or inaccessible"
	errMsgArchiveNotFound   = "archive not found"

	// Channel operation error messages
	errMsgCreateChannel       = "unable to create channel - ensure the target version exists"
	errMsgUpdateChannel       = "unable to update channel - ensure the target version exists"
	errMsgRetrieveChannel     = "unable to retrieve channel information"
	errMsgDeleteChannel       = "unable to delete channel"
	errMsgRetrieveChannelList = "unable to retrieve channel list for resource"
	errMsgChannelNotFound     = "channel not found"
	errMsgChannelExists       = "channel already exists"
)

// Implements the [Registry] interface using SQL databases.
//
// Provides persistent storage of namespaces, resources, versions, channels,
// and archives. Supports any SQL database with a compatible driver (SQLite,
// PostgreSQL, MySQL, etc.). Uses SQL for ACID transactions and referential
// integrity. Thread-safe for concurrent access. The registry does not own
// the database connection. The caller is responsible for connection lifecycle
// management, including calling Close() on the *sql.DB.
type SQLRegistry struct {
	db          *sql.DB      // Database connection
	logger      *slog.Logger // Logger for registry operations
	archiveRoot string       // Root directory for archive storage
	mu          sync.RWMutex // Protects archive storage operations
}

// Creates a new SQL database-backed registry.
//
// The caller is responsible for:
//   - Opening the database connection with appropriate driver and settings
//   - Enabling driver-specific features (e.g., PRAGMA foreign_keys for SQLite)
//   - Managing connection lifecycle (calling Close() when done)
//   - Providing the archiveRoot directory where archive files will be stored
//
// The registry will create the necessary schema if it doesn't exist.
func NewSQLRegistry(ctx context.Context, db *sql.DB, archiveRoot string, logger *slog.Logger) (*SQLRegistry, error) {
	if logger == nil {
		logger = slog.Default()
	}

	if _, err := db.ExecContext(ctx, sqlSchema); err != nil {
		logger.Error("failed to create schema", "error", err)
		return nil, &Error{
			Code:    ErrorCodeInternalError,
			Message: "failed to create schema",
		}
	}

	return &SQLRegistry{
		db:          db,
		logger:      logger,
		archiveRoot: archiveRoot,
	}, nil
}

// Creates a new namespace in the registry.
//
// Validates the namespace name before creation. If a namespace with the given
// name already exists, [ErrorCodeNamespaceExists] is returned. The response
// includes the created namespace's metadata with an empty resources list.
func (r *SQLRegistry) CreateNamespace(ctx context.Context, info NamespaceInfo) (ns *Namespace, err error) {
	if err := validateNamespace(info.Name); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	if ns, err = r.insertNamespace(ctx, info); err != nil {

		// Check if namespace already exists (only on error). There's a race
		// condition here, but the semantics hold: if it exists now, then it
		// already exists.
		if _, err := r.getNamespace(ctx, info.Name); err == nil {
			return nil, &Error{
				Code:    ErrorCodeNamespaceExists,
				Message: errMsgNamespaceExists,
			}
		}

		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgCreateNamespace, err, "namespace", info.Name)
	}

	return ns, nil
}

// Retrieves a namespace with resource summaries.
//
// Returns [ErrorCodeNotFound] if the namespace does not exist. The response
// includes namespace information along with lightweight summaries of all
// contained resources. Summaries include basic metadata (such as latest
// versions) but exclude full version histories.
func (r *SQLRegistry) ReadNamespace(ctx context.Context, namespace string) (*Namespace, error) {
	if err := validateNamespace(namespace); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	ns, err := r.getNamespace(ctx, namespace)
	if err == sql.ErrNoRows {
		return nil, &Error{Code: ErrorCodeNotFound, Message: errMsgNamespaceNotFound}
	}
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveNamespace, err, "namespace", namespace)
	}

	// Get resource summaries
	resources, err := r.listResources(ctx, namespace)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveResourceList, err, "namespace", namespace)
	}
	ns.Resources = resources

	return ns, nil
}

// Updates a namespace's mutable metadata.
//
// Only the description field can be modified. The namespace name cannot be changed
// after creation. Returns [ErrorCodeNotFound] if the namespace does not exist.
func (r *SQLRegistry) UpdateNamespace(ctx context.Context, namespace string, info NamespaceInfo) (*Namespace, error) {
	if err := validateNamespace(namespace); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	ns, err := r.updateNamespace(ctx, namespace, info)
	if err == sql.ErrNoRows {
		return nil, &Error{Code: ErrorCodeNotFound, Message: errMsgNamespaceNotFound}
	}
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgSaveNamespaceChanges, err, "namespace", namespace)
	}

	// Get resources for the namespace
	resources, err := r.listResources(ctx, namespace)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveResourceList, err, "namespace", namespace)
	}
	ns.Resources = resources

	return ns, nil
}

// Permanently deletes a namespace.
//
// Foreign key constraints prevent deletion if the namespace contains resources.
// The operation will fail with [ErrorCodeNamespaceNotEmpty] if resources exist.
func (r *SQLRegistry) DeleteNamespace(ctx context.Context, namespace string) error {
	if err := validateNamespace(namespace); err != nil {
		return &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	if err := r.deleteNamespace(ctx, namespace); err != nil {
		return r.logAndReturnError(ErrorCodeInternalError, errMsgDeleteNamespace, err, "namespace", namespace)
	}

	return nil
}

// Returns all namespaces in the registry.
//
// Returns a [NamespaceList] containing summary information for all namespaces,
// including resource counts. The list may be empty if no namespaces exist.
func (r *SQLRegistry) ListNamespaces(ctx context.Context) (*NamespaceList, error) {
	namespaces, err := r.listNamespaces(ctx)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveNamespaceList, err)
	}
	return &NamespaceList{Namespaces: namespaces}, nil
}

// Creates a new resource in a namespace.
//
// Returns [ErrorCodeResourceExists] if a resource with the same name already
// exists in the namespace. Returns [ErrorCodeNotFound] if the parent namespace
// does not exist.
func (r *SQLRegistry) CreateResource(ctx context.Context, namespace string, info ResourceInfo) (*Resource, error) {
	if err := validateIdentifier(namespace, info.Name); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	res, err := r.insertResource(ctx, namespace, info)
	if err != nil {

		// Check whether resource already exists (only on error)
		if _, err := r.getResource(ctx, namespace, info.Name); err == nil {
			return nil, &Error{Code: ErrorCodeResourceExists, Message: errMsgResourceExists}
		}

		// Check whether namespace exists (foreign key constraint - only on error)
		if _, err := r.getNamespace(ctx, namespace); err == sql.ErrNoRows {
			return nil, &Error{Code: ErrorCodeNotFound, Message: errMsgNamespaceNotFound}
		}

		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgCreateResource, err, "namespace", namespace, "resource", info.Name)
	}

	return res, nil
}

// Retrieves a resource with version and channel summaries.
//
// Returns [ErrorCodeNotFound] if the resource does not exist. The response
// includes lightweight summaries of all versions and channels for navigation.
func (r *SQLRegistry) ReadResource(ctx context.Context, namespace string, resource string) (*Resource, error) {
	if err := validateIdentifier(namespace, resource); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	res, err := r.getResource(ctx, namespace, resource)
	if err == sql.ErrNoRows {
		return nil, &Error{Code: ErrorCodeNotFound, Message: errMsgResourceNotFound}
	}
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveResource, err, "namespace", namespace)
	}

	// Get version summaries
	versions, err := r.listVersions(ctx, namespace, resource)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveVersionList, err, "namespace", namespace, "resource", resource)
	}
	res.Versions = versions

	// Get channel summaries
	channels, err := r.listChannels(ctx, namespace, resource)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveChannelList, err, "namespace", namespace, "resource", resource)
	}
	res.Channels = channels

	return res, nil
}

// Updates a resource's mutable metadata.
//
// The type and description fields can be modified. The resource name cannot be
// changed after creation. Returns [ErrorCodeNotFound] if the resource does not exist.
func (r *SQLRegistry) UpdateResource(ctx context.Context, namespace string, resource string, info ResourceInfo) (*Resource, error) {
	if err := validateIdentifier(namespace, resource); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	res, err := r.updateResource(ctx, namespace, resource, info)
	if err == sql.ErrNoRows {
		return nil, &Error{Code: ErrorCodeNotFound, Message: errMsgResourceNotFound}
	}
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgSaveResourceChanges, err, "namespace", namespace, "resource", resource)
	}

	// Get versions and channels for the resource
	versions, err := r.listVersions(ctx, namespace, resource)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveVersionList, err, "namespace", namespace, "resource", resource)
	}
	res.Versions = versions

	channels, err := r.listChannels(ctx, namespace, resource)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveChannelList, err, "namespace", namespace, "resource", resource)
	}
	res.Channels = channels

	return res, nil
}

// Permanently deletes a resource.
//
// Foreign key constraints prevent deletion if the resource contains versions.
// All versions and channels must be deleted first. This operation cannot be undone.
func (r *SQLRegistry) DeleteResource(ctx context.Context, namespace string, resource string) error {
	if err := validateIdentifier(namespace, resource); err != nil {
		return &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	if err := r.deleteResource(ctx, namespace, resource); err != nil {
		return r.logAndReturnError(ErrorCodeInternalError, errMsgDeleteResource, err, "namespace", namespace, "resource", resource)
	}

	return nil
}

// Returns all resources in a namespace.
//
// Returns a [ResourceList] containing summary information for all resources in
// the namespace, including version and channel counts. The list may be empty
// if no resources exist in the namespace.
func (r *SQLRegistry) ListResources(ctx context.Context, namespace string) (*ResourceList, error) {
	if err := validateNamespace(namespace); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	resources, err := r.listResources(ctx, namespace)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveResourceList, err, "namespace", namespace)
	}

	return &ResourceList{Resources: resources}, nil
}

// Creates a new version for a resource.
//
// Returns [ErrorCodeVersionExists] if a version with the same version string
// already exists for the resource. Returns [ErrorCodeNotFound] if the parent
// resource does not exist. Archives can be uploaded to the version after creation.
func (r *SQLRegistry) CreateVersion(ctx context.Context, namespace string, resource string, info VersionInfo) (*Version, error) {
	if err := validateReference(namespace, resource, info.String); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	// Try to insert the version
	v, err := r.insertVersion(ctx, namespace, resource, info)
	if err == nil {
		return v, nil
	}

	// Check if version already exists
	if _, checkErr := r.getVersion(ctx, namespace, resource, info.String); checkErr == nil {
		return nil, &Error{Code: ErrorCodeVersionExists, Message: errMsgVersionExists}
	}

	// Check if resource exists
	if _, checkErr := r.getResource(ctx, namespace, resource); checkErr == sql.ErrNoRows {
		return nil, &Error{Code: ErrorCodeNotFound, Message: errMsgResourceNotFound}
	}

	return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgCreateVersion, err, "namespace", namespace, "resource", resource, "version", info.String)
}

// Retrieves a version with its archive details.
//
// Returns [ErrorCodeNotFound] if the version does not exist. If an archive has
// been uploaded, the response includes the archive digest, size, and download URL.
func (r *SQLRegistry) ReadVersion(ctx context.Context, namespace string, resource string, version string) (*Version, error) {
	if err := validateReference(namespace, resource, version); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	v, err := r.getVersion(ctx, namespace, resource, version)
	if err == sql.ErrNoRows {
		return nil, &Error{Code: ErrorCodeNotFound, Message: errMsgVersionNotFound}
	}
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveVersion, err, "namespace", namespace, "resource", resource, "version", version)
	}
	return v, nil
}

// Updates a version's mutable metadata.
//
// Metadata updates are allowed even after publication to support documentation
// changes. The version string cannot be changed after creation. Returns
// [ErrorCodeNotFound] if the version does not exist.
func (r *SQLRegistry) UpdateVersion(ctx context.Context, namespace string, resource string, version string, info VersionInfo) (*Version, error) {
	if err := validateReference(namespace, resource, version); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	v, err := r.updateVersion(ctx, namespace, resource, version)
	if err == sql.ErrNoRows {
		return nil, &Error{Code: ErrorCodeNotFound, Message: errMsgVersionNotFound}
	}
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgSaveVersionChanges, err, "namespace", namespace, "resource", resource, "version", version)
	}
	return v, nil
}

// Permanently deletes a version.
//
// Foreign key constraints prevent deletion if the version is referenced by
// channels or has an associated archive. These must be deleted first. This
// operation cannot be undone.
func (r *SQLRegistry) DeleteVersion(ctx context.Context, namespace string, resource string, version string) error {
	if err := validateReference(namespace, resource, version); err != nil {
		return &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	if err := r.deleteVersion(ctx, namespace, resource, version); err != nil {
		return r.logAndReturnError(ErrorCodeInternalError, errMsgDeleteVersion, err, "namespace", namespace, "resource", resource, "version", version)
	}
	return nil
}

// Returns all versions for a resource.
//
// Returns a [VersionList] containing summary information for all versions of
// the resource. The list may be empty if no versions exist for the resource.
func (r *SQLRegistry) ListVersions(ctx context.Context, namespace string, resource string) (*VersionList, error) {
	if err := validateIdentifier(namespace, resource); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	versions, err := r.listVersions(ctx, namespace, resource)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveVersionList, err, "namespace", namespace, "resource", resource)
	}
	return &VersionList{Versions: versions}, nil
}

// Uploads an archive for a version.
//
// The archive data is hashed using SHA-256 to calculate the digest for content
// verification. If an archive already exists for the version, it will be replaced.
// Returns the updated version with populated archive metadata.
func (r *SQLRegistry) UploadArchive(ctx context.Context, namespace string, resource string, version string, archiveReader io.Reader) (*Version, error) {
	if err := validateReference(namespace, resource, version); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Store archive file and calculate digest
	digest, archivePath, size, err := r.storeArchiveFile(namespace, resource, version, archiveReader)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgStoreArchive, err, "namespace", namespace, "resource", resource, "version", version)
	}

	// Update version with archive metadata
	if err := r.uploadArchive(ctx, namespace, resource, version, digest, archivePath, size); err != nil {
		os.Remove(archivePath)
		if err == sql.ErrNoRows {
			return nil, r.logAndReturnError(ErrorCodeNotFound, errMsgVersionNotFound, err, "namespace", namespace, "resource", resource, "version", version)
		}
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgUpdateArchive, err, "namespace", namespace, "resource", resource, "version", version)
	}

	// Return the updated version with archive details
	v, err := r.getVersion(ctx, namespace, resource, version)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveVersion, err, "namespace", namespace, "resource", resource, "version", version)
	}

	return v, nil
}

// Returns a reader for a version's archive.
//
// Returns [ErrorCodeNotFound] if the version does not exist or if no archive
// has been uploaded for the version. The caller is responsible for closing the
// returned reader.
func (r *SQLRegistry) DownloadArchive(ctx context.Context, namespace string, resource string, version string) (io.ReadCloser, error) {
	if err := validateReference(namespace, resource, version); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	v, err := r.getVersion(ctx, namespace, resource, version)
	if err == sql.ErrNoRows {
		return nil, &Error{Code: ErrorCodeNotFound, Message: errMsgVersionNotFound}
	}
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveVersion, err, "namespace", namespace, "resource", resource, "version", version)
	}

	// Check if archive has been uploaded
	if v.Archive == nil {
		return nil, &Error{Code: ErrorCodeNotFound, Message: errMsgArchiveNotFound}
	}

	// Open and return archive file
	file, err := os.Open(*v.Archive)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgAccessArchiveFile, err, "namespace", namespace, "resource", resource, "version", version)
	}

	return file, nil
}

// Creates a new channel.
//
// Returns [ErrorCodeChannelExists] if a channel with the same name already
// exists. Returns [ErrorCodeNotFound] if the target version does not exist.
// Channels provide mutable pointers to versions for dynamic references.
func (r *SQLRegistry) CreateChannel(ctx context.Context, namespace string, resource string, info ChannelInfo) (*Channel, error) {
	if err := validateChannelInfo(namespace, resource, info); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	if err := r.insertChannel(ctx, namespace, resource, info); err != nil {

		// Check if channel already exists
		if _, checkErr := r.getChannel(ctx, namespace, resource, info.Name); checkErr == nil {
			return nil, &Error{Code: ErrorCodeChannelExists, Message: errMsgChannelExists}
		}

		// Check if resource exists (foreign key constraint)
		if _, checkErr := r.getResource(ctx, namespace, resource); checkErr == sql.ErrNoRows {
			return nil, &Error{Code: ErrorCodeNotFound, Message: errMsgResourceNotFound}
		}

		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgCreateChannel, err, "namespace", namespace, "resource", resource, "channel", info.Name)
	}

	// Get the newly created channel
	c, err := r.getChannel(ctx, namespace, resource, info.Name)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveChannel, err, "namespace", namespace, "resource", resource, "channel", info.Name)
	}
	return c, nil
}

// Updates a channel's mutable metadata.
//
// The target version and description can be modified. The channel name cannot
// be changed after creation. Returns [ErrorCodeNotFound] if the channel does
// not exist.
func (r *SQLRegistry) UpdateChannel(ctx context.Context, namespace string, resource string, channel string, info ChannelInfo) (*Channel, error) {
	if err := validateChannelInfo(namespace, resource, info); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	c, err := r.updateChannel(ctx, namespace, resource, info)
	if err == sql.ErrNoRows {
		return nil, &Error{Code: ErrorCodeNotFound, Message: errMsgChannelNotFound}
	}
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgUpdateChannel, err, "namespace", namespace, "resource", resource, "channel", info.Name)
	}
	return c, nil
}

// Retrieves a channel with its version details.
//
// Returns [ErrorCodeNotFound] if the channel does not exist. The response includes
// the full version object that the channel currently points to, including archive
// details if available.
func (r *SQLRegistry) ReadChannel(ctx context.Context, namespace string, resource string, channel string) (*Channel, error) {
	if err := validateChannelReference(namespace, resource, channel); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	c, err := r.getChannel(ctx, namespace, resource, channel)
	if err == sql.ErrNoRows {
		return nil, &Error{Code: ErrorCodeNotFound, Message: errMsgChannelNotFound}
	}
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveChannel, err, "namespace", namespace, "resource", resource, "channel", channel)
	}
	return c, nil
}

// Permanently deletes a channel.
//
// The referenced version and its archive are not affected. This operation only
// removes the mutable pointer to the version.
func (r *SQLRegistry) DeleteChannel(ctx context.Context, namespace string, resource string, channel string) error {
	if err := validateChannelReference(namespace, resource, channel); err != nil {
		return &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	if err := r.deleteChannel(ctx, namespace, resource, channel); err != nil {
		return r.logAndReturnError(ErrorCodeInternalError, errMsgDeleteChannel, err, "namespace", namespace, "resource", resource, "channel", channel)
	}
	return nil
}

// Returns all channels for a resource.
//
// Returns a [ChannelList] containing summary information for all channels,
// including the version each channel currently points to. The list may be empty
// if no channels exist for the resource.
func (r *SQLRegistry) ListChannels(ctx context.Context, namespace string, resource string) (*ChannelList, error) {
	if err := validateIdentifier(namespace, resource); err != nil {
		return nil, &Error{Code: ErrorCodeBadRequest, Message: err.Error()}
	}

	channels, err := r.listChannels(ctx, namespace, resource)
	if err != nil {
		return nil, r.logAndReturnError(ErrorCodeInternalError, errMsgRetrieveChannelList, err, "namespace", namespace, "resource", resource)
	}
	return &ChannelList{Channels: channels}, nil
}
