-- Updates a version with archive metadata after upload.
--
-- Sets the digest, size, and path fields which are NULL until an archive is uploaded.
UPDATE versions 
SET digest = ?, size = ?, path = ?, updated_at = ? 
WHERE namespace = ? AND resource = ? AND string = ?;
