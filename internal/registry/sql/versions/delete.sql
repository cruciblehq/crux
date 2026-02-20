-- Deletes a version.
--
-- Fails with foreign key constraint violation if any channels or archives
-- reference this version. Channels and archives must be deleted first.
DELETE FROM versions
WHERE namespace = ? AND resource = ? AND string = ?;
