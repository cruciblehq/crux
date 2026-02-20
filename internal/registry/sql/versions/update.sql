-- Updates an existing version's mutable fields.
UPDATE versions
SET updated_at = ?
WHERE namespace = ? AND resource = ? AND string = ?;
