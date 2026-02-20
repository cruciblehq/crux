package registry

import (
	"context"
	"database/sql"
	"time"
)

// Executes an INSERT statement for a new namespace.
//
// Returns the created namespace with initialized timestamps on success, or the
// raw database error on failure without any translation or logging.
func (r *SQLRegistry) insertNamespace(ctx context.Context, info NamespaceInfo) (*Namespace, error) {
	now := time.Now().Unix()

	_, err := r.db.ExecContext(ctx, sqlNamespacesInsert,
		info.Name,
		info.Description,
		now, // created_at
		now, // updated_at
	)

	if err != nil {
		return nil, err
	}

	return &Namespace{
		Name:        info.Name,
		Description: info.Description,
		Resources:   []ResourceSummary{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Queries a namespace by name from the database.
//
// Returns sql.ErrNoRows if the namespace does not exist. Returns the namespace
// without resource summaries on success.
func (r *SQLRegistry) getNamespace(ctx context.Context, name string) (*Namespace, error) {

	var ns Namespace

	if err := r.db.QueryRowContext(ctx, sqlNamespacesGet, name).Scan(
		&ns.Name,
		&ns.Description,
		&ns.CreatedAt,
		&ns.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &ns, nil
}

// Executes an UPDATE statement for a namespace's mutable fields.
//
// Returns the updated namespace on success, sql.ErrNoRows if the namespace does
// not exist, or the raw database error on failure without any translation or logging.
func (r *SQLRegistry) updateNamespace(ctx context.Context, namespace string, info NamespaceInfo) (*Namespace, error) {
	now := time.Now().Unix()

	result, err := r.db.ExecContext(ctx, sqlNamespacesUpdate, info.Description, now, namespace)
	if err != nil {
		return nil, err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return nil, sql.ErrNoRows
	}

	// Return updated namespace - need to query for created_at. We could use a
	// RETURNING clause, but the cost would be to lose compatibility with some
	// databases, like MySQL.
	return r.getNamespace(ctx, namespace)
}

// Executes a DELETE statement for a namespace.
//
// Returns the raw database error on failure without any translation or logging.
// Foreign key constraints prevent deletion if the namespace contains resources.
func (r *SQLRegistry) deleteNamespace(ctx context.Context, namespace string) error {
	_, err := r.db.ExecContext(ctx, sqlNamespacesDelete, namespace)
	return err
}

// Executes an INSERT statement for a new resource.
//
// Returns the created resource with initialized timestamps on success, or the
// raw database error on failure without any translation or logging.
func (r *SQLRegistry) insertResource(ctx context.Context, namespace string, info ResourceInfo) (*Resource, error) {
	now := time.Now().Unix()

	_, err := r.db.ExecContext(ctx, sqlResourcesInsert, namespace, info.Name, info.Type, info.Description, now, now)
	if err != nil {
		return nil, err
	}

	return &Resource{
		Namespace:   namespace,
		Name:        info.Name,
		Type:        info.Type,
		Description: info.Description,
		Versions:    []VersionSummary{},
		Channels:    []ChannelSummary{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Queries a resource by namespace and name from the database.
//
// Returns sql.ErrNoRows if the resource does not exist. Returns the resource
// without version and channel summaries on success.
func (r *SQLRegistry) getResource(ctx context.Context, namespace, resource string) (*Resource, error) {
	var res Resource
	var ns string

	err := r.db.QueryRowContext(ctx, sqlResourcesGet, namespace, resource).Scan(
		&ns, &res.Name, &res.Type, &res.Description, &res.CreatedAt, &res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	res.Namespace = namespace
	return &res, nil
}

// Executes an UPDATE statement for a resource's mutable fields.
//
// Returns the updated resource on success, sql.ErrNoRows if the resource does
// not exist, or the raw database error on failure without any translation or logging.
func (r *SQLRegistry) updateResource(ctx context.Context, namespace, resource string, info ResourceInfo) (*Resource, error) {
	now := time.Now().Unix()

	result, err := r.db.ExecContext(ctx, sqlResourcesUpdate, info.Type, info.Description, now, namespace, resource)
	if err != nil {
		return nil, err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return nil, sql.ErrNoRows
	}

	// Return updated resource - need to query for created_at.  We could use a
	// RETURNING clause, but the cost would be to lose compatibility with some
	// databases, like MySQL.
	return r.getResource(ctx, namespace, resource)
}

// Executes a DELETE statement for a resource.
//
// Returns the raw database error on failure without any translation or logging.
// Foreign key constraints prevent deletion if the resource contains versions.
func (r *SQLRegistry) deleteResource(ctx context.Context, namespace, resource string) error {
	_, err := r.db.ExecContext(ctx, sqlResourcesDelete, namespace, resource)
	return err
}

// Executes an INSERT statement for a new version.
//
// Returns the created version with initialized timestamps on success, or the
// raw database error on failure without any translation or logging. Archive
// fields (Digest, Size, Archive) are initialized to nil and must be populated
// via uploadArchive.
func (r *SQLRegistry) insertVersion(ctx context.Context, namespace, resource string, info VersionInfo) (*Version, error) {
	now := time.Now().Unix()

	_, err := r.db.ExecContext(ctx, sqlVersionsInsert, namespace, resource, info.String, now, now)
	if err != nil {
		return nil, err
	}

	return &Version{
		Namespace: namespace,
		Resource:  resource,
		String:    info.String,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Queries a version by namespace, resource, and version string from the database.
//
// Returns sql.ErrNoRows if the version does not exist. Archive fields (Digest, Size,
// Archive) are populated if an archive has been uploaded, otherwise they remain nil.
func (r *SQLRegistry) getVersion(ctx context.Context, namespace, resource, version string) (*Version, error) {
	var v Version
	var digest, path sql.NullString
	var size sql.NullInt64

	err := r.db.QueryRowContext(ctx, sqlVersionsGet, namespace, resource, version).Scan(
		&v.String, &digest, &size, &path, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	v.Namespace = namespace
	v.Resource = resource

	if digest.Valid {
		v.Digest = &digest.String
	}
	if size.Valid {
		v.Size = &size.Int64
	}
	if path.Valid {
		v.Archive = &path.String
	}

	return &v, nil
}

// Executes an UPDATE statement for a version's mutable fields.
//
// Returns the updated version on success, sql.ErrNoRows if the version does not
// exist, or the raw database error on failure without any translation or logging.
func (r *SQLRegistry) updateVersion(ctx context.Context, namespace, resource, version string) (*Version, error) {
	now := time.Now().Unix()

	result, err := r.db.ExecContext(ctx, sqlVersionsUpdate, now, namespace, resource, version)
	if err != nil {
		return nil, err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return nil, sql.ErrNoRows
	}

	// Return updated version - need to query for created_at and archive
	// details. We could use a RETURNING clause, but the cost would be to lose
	// compatibility with some databases, like MySQL.
	return r.getVersion(ctx, namespace, resource, version)
}

// Executes a DELETE statement for a version.
//
// Returns the raw database error on failure without any translation or logging.
// Foreign key constraints prevent deletion if the version is referenced by channels.
func (r *SQLRegistry) deleteVersion(ctx context.Context, namespace, resource, version string) error {
	_, err := r.db.ExecContext(ctx, sqlVersionsDelete, namespace, resource, version)
	return err
}

// Executes an INSERT statement for a channel.
//
// Returns the raw database error on failure without any translation or logging.
func (r *SQLRegistry) insertChannel(ctx context.Context, namespace, resource string, info ChannelInfo) error {
	now := time.Now().Unix()
	_, err := r.db.ExecContext(ctx, sqlChannelsInsert, namespace, resource, info.Name, info.Description, info.Version, now, now)
	return err
}

// Executes an UPDATE statement for a channel.
//
// Returns the updated channel on success, sql.ErrNoRows if the channel does not
// exist, or the raw database error on failure without any translation or logging.
func (r *SQLRegistry) updateChannel(ctx context.Context, namespace, resource string, info ChannelInfo) (*Channel, error) {
	now := time.Now().Unix()

	result, err := r.db.ExecContext(ctx, sqlChannelsUpdate, info.Description, info.Version, now, namespace, resource, info.Name)
	if err != nil {
		return nil, err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return nil, sql.ErrNoRows
	}

	// Return updated channel - need to query for created_at and version
	// details. We could use a RETURNING clause, but the cost would be to lose
	// compatibility with some databases, like MySQL.
	return r.getChannel(ctx, namespace, resource, info.Name)
}

// Queries a channel by namespace, resource, and name from the database.
//
// Returns sql.ErrNoRows if the channel does not exist. Joins with the versions
// table to include the full version details that the channel points to.
func (r *SQLRegistry) getChannel(ctx context.Context, namespace, resource, channel string) (*Channel, error) {
	var c Channel
	var versionString string
	var channelCreatedAt, channelUpdatedAt, versionCreatedAt, versionUpdatedAt int64
	var digest, size, path sql.NullString

	err := r.db.QueryRowContext(ctx, sqlChannelsGet, namespace, resource, channel).Scan(
		&c.Name, &c.Description, &versionString, &channelCreatedAt, &channelUpdatedAt,
		&versionCreatedAt, &versionUpdatedAt, &digest, &size, &path,
	)
	if err != nil {
		return nil, err
	}

	c.Namespace = namespace
	c.Resource = resource
	c.CreatedAt = channelCreatedAt
	c.UpdatedAt = channelUpdatedAt
	c.Version = Version{
		Namespace: namespace, Resource: resource, String: versionString,
		CreatedAt: versionCreatedAt, UpdatedAt: versionUpdatedAt,
	}
	if digest.Valid {
		c.Version.Digest = &digest.String
	}
	return &c, nil
}

// Executes a DELETE statement for a channel.
//
// Returns the raw database error on failure without any translation or logging.
func (r *SQLRegistry) deleteChannel(ctx context.Context, namespace, resource, channel string) error {
	_, err := r.db.ExecContext(ctx, sqlChannelsDelete, namespace, resource, channel)
	return err
}

// Queries all namespaces from the database.
//
// Returns the raw database error on failure without any translation or logging.
func (r *SQLRegistry) listNamespaces(ctx context.Context) ([]NamespaceSummary, error) {
	rows, err := r.db.QueryContext(ctx, sqlNamespacesList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var namespaces []NamespaceSummary
	for rows.Next() {
		var ns NamespaceSummary
		if err := rows.Scan(&ns.Name, &ns.Description, &ns.CreatedAt, &ns.ResourceCount); err != nil {
			return nil, err
		}
		namespaces = append(namespaces, ns)
	}
	return namespaces, rows.Err()
}

// Queries all resources in a namespace from the database.
//
// Returns the raw database error on failure without any translation or logging.
func (r *SQLRegistry) listResources(ctx context.Context, namespace string) ([]ResourceSummary, error) {
	rows, err := r.db.QueryContext(ctx, sqlResourcesList, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []ResourceSummary
	for rows.Next() {
		var rs ResourceSummary
		var latestVersion sql.NullString
		if err := rows.Scan(&rs.Name, &rs.Type, &rs.Description, &rs.CreatedAt, &rs.UpdatedAt, &rs.VersionCount, &rs.ChannelCount, &latestVersion); err != nil {
			return nil, err
		}
		if latestVersion.Valid {
			rs.LatestVersion = &latestVersion.String
		}
		resources = append(resources, rs)
	}
	return resources, rows.Err()
}

// Queries all versions for a resource from the database.
//
// Returns the raw database error on failure without any translation or logging.
func (r *SQLRegistry) listVersions(ctx context.Context, namespace, resource string) ([]VersionSummary, error) {
	rows, err := r.db.QueryContext(ctx, sqlVersionsList, namespace, resource)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []VersionSummary
	for rows.Next() {
		var vs VersionSummary
		var digest, size sql.NullString
		if err := rows.Scan(&vs.String, &vs.CreatedAt, &vs.UpdatedAt, &digest, &size); err != nil {
			return nil, err
		}
		versions = append(versions, vs)
	}
	return versions, rows.Err()
}

// Queries all channels for a resource from the database.
//
// Returns the raw database error on failure without any translation or logging.
func (r *SQLRegistry) listChannels(ctx context.Context, namespace, resource string) ([]ChannelSummary, error) {
	rows, err := r.db.QueryContext(ctx, sqlChannelsList, namespace, resource)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []ChannelSummary
	for rows.Next() {
		var cs ChannelSummary
		if err := rows.Scan(&cs.Name, &cs.Description, &cs.Version, &cs.CreatedAt, &cs.UpdatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, cs)
	}
	return channels, rows.Err()
}

// Executes an UPDATE statement to set archive metadata for a version.
//
// Returns sql.ErrNoRows if the version does not exist, or the raw database
// error on failure without any translation or logging.
func (r *SQLRegistry) uploadArchive(ctx context.Context, namespace, resource, version, digest, path string, size int64) error {
	now := time.Now().Unix()

	result, err := r.db.ExecContext(ctx, sqlVersionsUpload, digest, size, path, now, namespace, resource, version)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
