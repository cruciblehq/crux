package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Returns cached namespace info and ETag, or ErrNotFound.
//
// The ETag can be used for conditional requests with If-None-Match. If the
// caller only needs the ETag, they can discard the NamespaceInfo.
func (c *Cache) GetNamespace(namespace string) (*NamespaceInfo, string, error) {
	var description, etag string

	err := c.db.QueryRow(sqlNamespacesGet, namespace).Scan(&description, &etag)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", fmt.Errorf("namespace %s: %w", namespace, ErrNotFound)
	}
	if err != nil {
		return nil, "", err
	}

	summaries, err := c.getResourceSummaries(namespace)
	if err != nil {
		return nil, "", err
	}

	return &NamespaceInfo{
		Namespace:   namespace,
		Description: description,
		Resources:   summaries,
	}, etag, nil
}

// Replaces cached namespace and its resource summaries.
//
// Deletes any existing data for the namespace and inserts the new data
// atomically within a transaction. The delete cascades to resource_summaries
// via the foreign key constraint.
func (c *Cache) PutNamespace(info *NamespaceInfo, etag string) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().Unix()

	// Cascades to resource_summaries.
	_, err = tx.Exec(sqlNamespacesDelete, info.Namespace)
	if err != nil {
		return err
	}

	_, err = tx.Exec(sqlNamespacesInsert,
		info.Namespace,
		info.Description,
		etag,
		now,
		now,
	)
	if err != nil {
		return err
	}

	for _, r := range info.Resources {
		_, err = tx.Exec(sqlResourceSummariesInsert,
			info.Namespace,
			r.Name,
			r.Type,
			r.Description,
			r.Latest,
			r.UpdatedAt.Unix(),
			now,
			now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Fetches resource summaries for a namespace.
func (c *Cache) getResourceSummaries(namespace string) ([]ResourceSummary, error) {
	rows, err := c.db.Query(sqlResourceSummariesList, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []ResourceSummary
	for rows.Next() {
		var s ResourceSummary
		var updatedAt int64

		err := rows.Scan(
			&s.Name,
			&s.Type,
			&s.Description,
			&s.Latest,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		s.UpdatedAt = time.Unix(updatedAt, 0)
		summaries = append(summaries, s)
	}

	return summaries, rows.Err()
}
