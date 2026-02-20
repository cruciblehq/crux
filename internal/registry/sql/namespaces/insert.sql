-- Creates a new namespace.
--
-- The namespace contains no resources upon creation.
INSERT INTO namespaces (name, description, created_at, updated_at)
VALUES (?, ?, ?, ?);
