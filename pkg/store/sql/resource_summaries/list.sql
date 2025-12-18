SELECT name, type, description, latest, updated_at
FROM resource_summaries
WHERE namespace = ?;
