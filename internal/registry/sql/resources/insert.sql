-- Creates a new resource.
--
-- The resource is empty, containing no versions upon creation.
INSERT INTO resources (namespace, name, type, description, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);
