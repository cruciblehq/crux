-- Cache database schema.
--
-- Stores metadata from registry endpoints and tracks locally extracted archives.
-- Each table group corresponds to a specific registry endpoint:
--
--   namespaces, resource_summaries  <- GET /v1/{namespace}
--   resources, versions, channels   <- GET /v1/{namespace}/{name}
--   archives                        <- GET /v1/{namespace}/{name}/{ref}

-- Namespace metadata from GET /v1/{namespace}.
CREATE TABLE IF NOT EXISTS namespaces (
    namespace   TEXT NOT NULL,        -- Namespace identifier.
    description TEXT NOT NULL,        -- Human-readable description.
    etag        TEXT NOT NULL,        -- ETag for cache validation.
    created_at  INTEGER NOT NULL,     -- Unix timestamp when first cached.
    updated_at  INTEGER NOT NULL,     -- Unix timestamp when last updated.
    PRIMARY KEY (namespace)
);

-- Resource summaries within a namespace, from GET /v1/{namespace}. These are
-- lightweight listings, not full resource metadata.
CREATE TABLE IF NOT EXISTS resource_summaries (
    namespace   TEXT NOT NULL,        -- Parent namespace.
    name        TEXT NOT NULL,        -- Resource name.
    type        TEXT NOT NULL,        -- Resource type.
    description TEXT NOT NULL,        -- Human-readable description.
    latest      TEXT NOT NULL,        -- Latest stable version string.
    created_at  INTEGER NOT NULL,     -- Unix timestamp when first cached.
    updated_at  INTEGER NOT NULL,     -- Unix timestamp when last updated.
    PRIMARY KEY (namespace, name),
    FOREIGN KEY (namespace) REFERENCES namespaces (namespace) ON DELETE CASCADE
);

-- Full resource metadata from GET /v1/{namespace}/{name}.
CREATE TABLE IF NOT EXISTS resources (
    namespace   TEXT NOT NULL,        -- Parent namespace.
    name        TEXT NOT NULL,        -- Resource name.
    type        TEXT NOT NULL,        -- Resource type.
    description TEXT NOT NULL,        -- Human-readable description.
    etag        TEXT NOT NULL,        -- ETag for cache validation.
    created_at  INTEGER NOT NULL,     -- Unix timestamp when first cached.
    updated_at  INTEGER NOT NULL,     -- Unix timestamp when last updated.
    PRIMARY KEY (namespace, name)
);

-- Available versions for a resource, from GET /v1/{namespace}/{name}.
CREATE TABLE IF NOT EXISTS versions (
    namespace   TEXT NOT NULL,        -- Parent namespace.
    name        TEXT NOT NULL,        -- Parent resource name.
    version     TEXT NOT NULL,        -- Semantic version string.
    digest      TEXT NOT NULL,        -- Content digest for integrity verification.
    published   INTEGER NOT NULL,     -- Unix timestamp when published.
    size        INTEGER NOT NULL,     -- Archive size in bytes.
    created_at  INTEGER NOT NULL,     -- Unix timestamp when first cached.
    updated_at  INTEGER NOT NULL,     -- Unix timestamp when last updated.
    PRIMARY KEY (namespace, name, version),
    FOREIGN KEY (namespace, name) REFERENCES resources (namespace, name) ON DELETE CASCADE
);

-- Release channels for a resource, from GET /v1/{namespace}/{name}. Channels
-- are named pointers to specific versions.
CREATE TABLE IF NOT EXISTS channels (
    namespace   TEXT NOT NULL,        -- Parent namespace.
    name        TEXT NOT NULL,        -- Parent resource name.
    channel     TEXT NOT NULL,        -- Channel name.
    description TEXT NOT NULL,        -- Human-readable description.
    version     TEXT NOT NULL,        -- Version this channel points to.
    digest      TEXT NOT NULL,        -- Content digest for integrity verification.
    published   INTEGER NOT NULL,     -- Unix timestamp when version was published.
    size        INTEGER NOT NULL,     -- Archive size in bytes.
    created_at  INTEGER NOT NULL,     -- Unix timestamp when first cached.
    updated_at  INTEGER NOT NULL,     -- Unix timestamp when last updated.
    PRIMARY KEY (namespace, name, channel),
    FOREIGN KEY (namespace, name) REFERENCES resources (namespace, name) ON DELETE CASCADE
);

-- Locally extracted archives from GET /v1/{namespace}/{name}/{ref}. Tracks what
-- has been downloaded and where it lives on disk. Independent of other tables;
-- an archive may exist even if metadata is stale.
CREATE TABLE IF NOT EXISTS archives (
    namespace   TEXT NOT NULL,        -- Parent namespace.
    name        TEXT NOT NULL,        -- Parent resource name.
    version     TEXT NOT NULL,        -- Semantic version string.
    digest      TEXT NOT NULL,        -- Content digest for integrity verification.
    path        TEXT NOT NULL,        -- Filesystem path to extracted directory.
    etag        TEXT NOT NULL,        -- ETag for cache validation.
    created_at  INTEGER NOT NULL,     -- Unix timestamp when first cached.
    updated_at  INTEGER NOT NULL,     -- Unix timestamp when last updated.
    PRIMARY KEY (namespace, name, version)
);
