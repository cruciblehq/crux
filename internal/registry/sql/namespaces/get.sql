-- Retrieves a single namespace.
SELECT 
    name,
    description,
    created_at,
    updated_at
FROM namespaces
WHERE name = ?;
