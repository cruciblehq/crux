-- Lists all namespaces with their resource counts.
--
-- Returns namespace metadata and counts of resources within each namespace.
SELECT 
    namespaces.name,
    namespaces.description,
    namespaces.created_at,
    COUNT(resources.name) as resource_count
FROM namespaces
LEFT JOIN resources ON resources.namespace = namespaces.name
GROUP BY namespaces.name, namespaces.description, namespaces.created_at
ORDER BY namespaces.name ASC;
