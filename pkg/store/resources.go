package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cruciblehq/crux/pkg/reference"
)

// Returns cached resource info and ETag, or ErrNotFound.
//
// The ETag can be used for conditional requests with If-None-Match. If the
// caller only needs the ETag, they can discard the ResourceInfo.
func (c *Cache) GetResource(namespace, name string) (*ResourceInfo, string, error) {
	var resourceType, description, etag string

	err := c.db.QueryRow(sqlResourcesGet, namespace, name).Scan(&resourceType, &description, &etag)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", fmt.Errorf("resource %s/%s: %w", namespace, name, ErrNotFound)
	}
	if err != nil {
		return nil, "", err
	}

	versions, err := c.getVersions(namespace, name)
	if err != nil {
		return nil, "", err
	}

	channels, err := c.getChannels(namespace, name)
	if err != nil {
		return nil, "", err
	}

	return &ResourceInfo{
		Name:        name,
		Type:        resourceType,
		Description: description,
		Versions:    versions,
		Channels:    channels,
	}, etag, nil
}

// Replaces cached resource and its versions/channels.
//
// Deletes any existing data for the resource and inserts the new data
// atomically within a transaction. The delete cascades to versions and
// channels via foreign key constraints.
func (c *Cache) PutResource(namespace string, info *ResourceInfo, etag string) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().Unix()

	// Cascades to versions and channels.
	_, err = tx.Exec(sqlResourcesDelete, namespace, info.Name)
	if err != nil {
		return err
	}

	_, err = tx.Exec(sqlResourcesInsert,
		namespace,
		info.Name,
		info.Type,
		info.Description,
		etag,
		now,
		now,
	)
	if err != nil {
		return err
	}

	for _, v := range info.Versions {
		_, err = tx.Exec(sqlVersionsInsert,
			namespace,
			info.Name,
			v.Version.String(),
			v.Digest.String(),
			v.Published.Unix(),
			v.Size,
			now,
			now,
		)
		if err != nil {
			return err
		}
	}

	for _, ch := range info.Channels {
		_, err = tx.Exec(sqlChannelsInsert,
			namespace,
			info.Name,
			ch.Channel,
			ch.Description,
			ch.Version.String(),
			ch.Digest.String(),
			ch.Published.Unix(),
			ch.Size,
			now,
			now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Fetches versions for a resource.
func (c *Cache) getVersions(namespace, name string) ([]VersionInfo, error) {
	rows, err := c.db.Query(sqlVersionsList, namespace, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []VersionInfo
	for rows.Next() {
		var versionStr, digestStr string
		var published, size int64

		err := rows.Scan(&versionStr, &digestStr, &published, &size)
		if err != nil {
			return nil, err
		}

		ver, err := reference.ParseVersion(versionStr)
		if err != nil {
			return nil, err
		}

		digest, err := reference.ParseDigest(digestStr)
		if err != nil {
			return nil, err
		}

		versions = append(versions, VersionInfo{
			Version:   *ver,
			Digest:    *digest,
			Published: time.Unix(published, 0),
			Size:      size,
		})
	}

	return versions, rows.Err()
}

// Fetches channels for a resource.
func (c *Cache) getChannels(namespace, name string) ([]ChannelInfo, error) {
	rows, err := c.db.Query(sqlChannelsList, namespace, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []ChannelInfo
	for rows.Next() {
		var channel, description, versionStr, digestStr string
		var published, size int64

		err := rows.Scan(&channel, &description, &versionStr, &digestStr, &published, &size)
		if err != nil {
			return nil, err
		}

		ver, err := reference.ParseVersion(versionStr)
		if err != nil {
			return nil, err
		}

		digest, err := reference.ParseDigest(digestStr)
		if err != nil {
			return nil, err
		}

		channels = append(channels, ChannelInfo{
			VersionInfo: VersionInfo{
				Version:   *ver,
				Digest:    *digest,
				Published: time.Unix(published, 0),
				Size:      size,
			},
			Channel:     channel,
			Description: description,
		})
	}

	return channels, rows.Err()
}
