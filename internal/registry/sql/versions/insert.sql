-- Creates a new version.
--
-- The version is created without an archive (digest, size, path are NULL),
-- which must be uploaded separately using UploadArchive.
INSERT INTO versions (namespace, resource, string, digest, size, path, created_at, updated_at)
VALUES (?, ?, ?, NULL, NULL, NULL, ?, ?);
