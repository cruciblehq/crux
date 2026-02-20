-- Retrieves resource metadata.
SELECT 
    namespace,
    name,
    type,
    description,
    created_at,
    updated_at
FROM resources
WHERE namespace = ? AND name = ?;
