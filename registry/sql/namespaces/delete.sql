-- Deletes a namespace.
--
-- Fails with foreign key constraint violation if the namespace contains any
-- resources. Resources must be deleted first.
DELETE FROM namespaces
WHERE name = ?;
