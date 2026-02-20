-- Updates an existing namespace's metadata.
UPDATE namespaces
SET description = ?, updated_at = ?
WHERE name = ?;
