-- Creates a new channel.
INSERT INTO channels (namespace, resource, name, description, version, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);
