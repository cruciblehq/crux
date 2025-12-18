package store

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/cruciblehq/crux/pkg/reference"
)

func TestGetArchive(t *testing.T) {
	c, err := openCacheWithPath(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening cache: %v", err)
	}
	defer c.Close()

	t.Run("not found", func(t *testing.T) {
		version, _ := reference.ParseVersion("1.0.0")
		_, err := c.GetArchive("crucible", "starter", *version)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("found", func(t *testing.T) {
		version, _ := reference.ParseVersion("1.0.0")
		digest, _ := reference.ParseDigest("sha256:abc123")

		archive := &Archive{
			Namespace: "crucible",
			Name:      "starter",
			Version:   *version,
			Digest:    *digest,
			Path:      "/path/to/extracted",
			ETag:      "etag-123",
		}

		err := c.PutArchive(archive)
		if err != nil {
			t.Fatalf("putting archive: %v", err)
		}

		got, err := c.GetArchive("crucible", "starter", *version)
		if err != nil {
			t.Fatalf("getting archive: %v", err)
		}

		if got.Namespace != archive.Namespace {
			t.Errorf("namespace: got %q, want %q", got.Namespace, archive.Namespace)
		}
		if got.Name != archive.Name {
			t.Errorf("name: got %q, want %q", got.Name, archive.Name)
		}
		if got.Version.String() != archive.Version.String() {
			t.Errorf("version: got %q, want %q", got.Version.String(), archive.Version.String())
		}
		if got.Digest.String() != archive.Digest.String() {
			t.Errorf("digest: got %q, want %q", got.Digest.String(), archive.Digest.String())
		}
		if got.Path != archive.Path {
			t.Errorf("path: got %q, want %q", got.Path, archive.Path)
		}
		if got.ETag != archive.ETag {
			t.Errorf("etag: got %q, want %q", got.ETag, archive.ETag)
		}
	})
}

func TestPutArchive(t *testing.T) {
	c, err := openCacheWithPath(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening cache: %v", err)
	}
	defer c.Close()

	version, _ := reference.ParseVersion("1.0.0")
	digest, _ := reference.ParseDigest("sha256:abc123")

	t.Run("insert", func(t *testing.T) {
		archive := &Archive{
			Namespace: "crucible",
			Name:      "starter",
			Version:   *version,
			Digest:    *digest,
			Path:      "/path/to/extracted",
			ETag:      "etag-123",
		}

		err := c.PutArchive(archive)
		if err != nil {
			t.Fatalf("putting archive: %v", err)
		}

		got, err := c.GetArchive("crucible", "starter", *version)
		if err != nil {
			t.Fatalf("getting archive: %v", err)
		}

		if got.Path != archive.Path {
			t.Errorf("path: got %q, want %q", got.Path, archive.Path)
		}
	})

	t.Run("duplicate fails", func(t *testing.T) {
		archive := &Archive{
			Namespace: "crucible",
			Name:      "starter",
			Version:   *version,
			Digest:    *digest,
			Path:      "/path/to/different",
			ETag:      "etag-456",
		}

		err := c.PutArchive(archive)
		if err == nil {
			t.Error("expected error on duplicate insert, got nil")
		}
	})
}

func TestDeleteArchive(t *testing.T) {
	c, err := openCacheWithPath(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening cache: %v", err)
	}
	defer c.Close()

	version, _ := reference.ParseVersion("1.0.0")
	digest, _ := reference.ParseDigest("sha256:abc123")

	archive := &Archive{
		Namespace: "crucible",
		Name:      "starter",
		Version:   *version,
		Digest:    *digest,
		Path:      "/path/to/extracted",
		ETag:      "etag-123",
	}

	err = c.PutArchive(archive)
	if err != nil {
		t.Fatalf("putting archive: %v", err)
	}

	err = c.DeleteArchive("crucible", "starter", *version)
	if err != nil {
		t.Fatalf("deleting archive: %v", err)
	}

	_, err = c.GetArchive("crucible", "starter", *version)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}
