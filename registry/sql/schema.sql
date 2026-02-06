CREATE TABLE IF NOT EXISTS namespaces (
    name        TEXT NOT NULL,        -- Namespace identifier.
    description TEXT NOT NULL,        -- Human-readable description.
    created_at  INTEGER NOT NULL,     -- Unix timestamp when first cached.
    updated_at  INTEGER NOT NULL,     -- Unix timestamp when last updated.
    PRIMARY KEY (name)
);

CREATE TABLE IF NOT EXISTS resources (
    namespace   TEXT NOT NULL,        -- Parent namespace.
    name        TEXT NOT NULL,        -- Resource name.
    type        TEXT NOT NULL,        -- Resource type.
    description TEXT NOT NULL,        -- Human-readable description.
    created_at  INTEGER NOT NULL,     -- Unix timestamp when first cached.
    updated_at  INTEGER NOT NULL,     -- Unix timestamp when last updated.
    PRIMARY KEY (namespace, name),
    FOREIGN KEY (namespace) REFERENCES namespaces(name) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS versions (
    namespace    TEXT NOT NULL,        -- Parent namespace.
    resource     TEXT NOT NULL,        -- Parent resource name.
    string       TEXT NOT NULL,        -- Semantic version string.
    digest       TEXT,                 -- Archive content digest (NULL until uploaded).
    size         INTEGER,              -- Archive size in bytes (NULL until uploaded).
    path         TEXT,                 -- Filesystem path to archive file (NULL until uploaded).
    created_at   INTEGER NOT NULL,     -- Unix timestamp when first cached.
    updated_at   INTEGER NOT NULL,     -- Unix timestamp when last updated.
    PRIMARY KEY (namespace, resource, string),
    FOREIGN KEY (namespace, resource) REFERENCES resources (namespace, name) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS channels (
    namespace   TEXT NOT NULL,        -- Parent namespace.
    resource    TEXT NOT NULL,        -- Parent resource name.
    name        TEXT NOT NULL,        -- Channel name.
    description TEXT NOT NULL,        -- Human-readable description.
    version     TEXT NOT NULL,        -- Version this channel points to.
    created_at  INTEGER NOT NULL,     -- Unix timestamp when first cached.
    updated_at  INTEGER NOT NULL,     -- Unix timestamp when last updated.
    PRIMARY KEY (namespace, resource, name),
    FOREIGN KEY (namespace, resource) REFERENCES resources (namespace, name) ON DELETE RESTRICT,
    FOREIGN KEY (namespace, resource, version) REFERENCES versions (namespace, resource, string) ON DELETE RESTRICT
);
