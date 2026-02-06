-- Lists all channels for a resource.
SELECT 
    name,
    description,
    version,
    created_at,
    updated_at
FROM channels
WHERE namespace = ? AND resource = ?
ORDER BY name;
