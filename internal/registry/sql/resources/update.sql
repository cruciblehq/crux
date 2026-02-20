-- Updates an existing resource's mutable fields.
UPDATE resources
SET type = ?, description = ?, updated_at = ?
WHERE namespace = ? AND name = ?;
