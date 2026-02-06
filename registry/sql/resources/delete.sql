-- Deletes a resource.
--
-- Fails with foreign key constraint violation if the resource contains any
-- versions. Versions must be deleted first.
DELETE FROM resources
WHERE namespace = ? AND name = ?;
