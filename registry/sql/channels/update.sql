-- Updates an existing channel to point to a different version.
UPDATE channels
SET description = ?, version = ?, updated_at = ?
WHERE namespace = ? AND resource = ? AND name = ?;
