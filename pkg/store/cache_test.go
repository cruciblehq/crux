package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen(t *testing.T) {
	t.Run("creates directory", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "subdir", "test.db")

		c, err := openCacheWithPath(dbPath)
		if err != nil {
			t.Fatalf("opening cache: %v", err)
		}
		defer c.Close()

		if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
			t.Error("expected directory to be created")
		}
	})

	t.Run("creates database file", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		c, err := openCacheWithPath(dbPath)
		if err != nil {
			t.Fatalf("opening cache: %v", err)
		}
		defer c.Close()

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Error("expected database file to be created")
		}
	})

	t.Run("reopens existing database", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		c1, err := openCacheWithPath(dbPath)
		if err != nil {
			t.Fatalf("opening cache: %v", err)
		}
		c1.Close()

		c2, err := openCacheWithPath(dbPath)
		if err != nil {
			t.Fatalf("reopening cache: %v", err)
		}
		defer c2.Close()
	})

	t.Run("invalid path", func(t *testing.T) {
		_, err := openCacheWithPath("/dev/null/invalid/path.db")
		if err == nil {
			t.Error("expected error for invalid path")
		}
	})
}

func TestClose(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	c, err := openCacheWithPath(dbPath)
	if err != nil {
		t.Fatalf("opening cache: %v", err)
	}

	err = c.Close()
	if err != nil {
		t.Errorf("closing cache: %v", err)
	}
}
