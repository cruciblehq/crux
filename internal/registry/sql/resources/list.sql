-- Lists all resources in a namespace with summary statistics.
--
-- Returns resource metadata with counts and latest version information. Used
-- for ResourceList responses and Namespace.resources field.
SELECT 
    resources.name,
    resources.type,
    resources.description,
    resources.created_at,
    resources.updated_at,
    COUNT(DISTINCT versions.string) as version_count,
    COUNT(DISTINCT channels.name) as channel_count,
    MAX(versions.string) as latest_version
FROM resources
LEFT JOIN versions ON versions.namespace = resources.namespace AND versions.resource = resources.name
LEFT JOIN channels ON channels.namespace = resources.namespace AND channels.resource = resources.name
WHERE resources.namespace = ?
GROUP BY resources.name, resources.type, resources.description, resources.created_at, resources.updated_at
ORDER BY resources.name;
