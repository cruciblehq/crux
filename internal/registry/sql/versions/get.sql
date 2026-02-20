-- Retrieves a specific version with its archive information.
--
-- Archive fields (digest, size, path) are NULL if no archive has been uploaded.
SELECT 
    string,
    digest,
    size,
    path,
    created_at,
    updated_at
FROM versions
WHERE namespace = ? AND resource = ? AND string = ?;
