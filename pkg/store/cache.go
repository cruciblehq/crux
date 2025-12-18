package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/cruciblehq/crux/pkg/paths"
)

const (

	// Default filename for the cache database, which is stored in the user's
	// cache directory.
	defaultCacheFileName = "cache.db"
)

// Provides local SQLite storage for registry metadata and archives.
type Cache struct {
	db *sql.DB
}

// Opens the cache database at the default path, creating it if necessary.
//
// The database is stored at ~/.cache/crux/cache.db (or platform equivalent).
// SQLite is configured with foreign_keys=ON to enforce foreign key constraints
// and journal_mode=WAL for concurrent reads.
func OpenCache() (*Cache, error) {
	dbPath := filepath.Join(paths.Cache(), defaultCacheFileName)
	return openCacheWithPath(dbPath)
}

// Opens the cache database at the given path.
//
// This is mostly useful for testing with a temporary database file. Creates
// the parent directory if it doesn't exist.
func openCacheWithPath(path string) (*Cache, error) {
	if err := os.MkdirAll(filepath.Dir(path), paths.DefaultDirMode); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	// Enable foreign keys (FK constraints) and WAL mode (concurrent reads).
	dsn := path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	c := &Cache{db: db}

	if err := c.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return c, nil
}

// Closes the cache database.
func (c *Cache) Close() error {
	return c.db.Close()
}

// Runs database migrations.
//
// Executes the full schema on every open. SQLite's IF NOT EXISTS clauses make
// this idempotent.
func (c *Cache) migrate() error {
	_, err := c.db.Exec(sqlSchema)
	return err
}
