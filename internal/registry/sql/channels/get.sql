-- Retrieves a channel's metadata along with its version and archive metadata.
SELECT 
    channels.name,
    channels.description,
    channels.version,
    channels.created_at,
    channels.updated_at,
    versions.created_at as version_created_at,
    versions.updated_at as version_updated_at,
    versions.digest,
    versions.size,
    versions.path
FROM channels
INNER JOIN versions ON versions.namespace = channels.namespace AND versions.resource = channels.resource AND versions.string = channels.version
WHERE channels.namespace = ? AND channels.resource = ? AND channels.name = ?;
