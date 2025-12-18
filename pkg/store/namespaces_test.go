package store

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestGetNamespace(t *testing.T) {
	c, err := openCacheWithPath(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening cache: %v", err)
	}
	defer c.Close()

	t.Run("not found", func(t *testing.T) {
		_, _, err := c.GetNamespace("crucible")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("found", func(t *testing.T) {
		info := &NamespaceInfo{
			Namespace:   "crucible",
			Description: "Official templates",
			Resources: []ResourceSummary{
				{
					Name:        "starter",
					Type:        "template",
					Description: "A starter template",
					Latest:      "1.0.0",
					UpdatedAt:   time.Unix(1000000, 0),
				},
			},
		}

		err := c.PutNamespace(info, "etag-123")
		if err != nil {
			t.Fatalf("putting namespace: %v", err)
		}

		got, etag, err := c.GetNamespace("crucible")
		if err != nil {
			t.Fatalf("getting namespace: %v", err)
		}

		if etag != "etag-123" {
			t.Errorf("etag: got %q, want %q", etag, "etag-123")
		}
		if got.Namespace != info.Namespace {
			t.Errorf("namespace: got %q, want %q", got.Namespace, info.Namespace)
		}
		if got.Description != info.Description {
			t.Errorf("description: got %q, want %q", got.Description, info.Description)
		}
		if len(got.Resources) != 1 {
			t.Fatalf("resources: got %d, want 1", len(got.Resources))
		}
		if got.Resources[0].Name != "starter" {
			t.Errorf("resource name: got %q, want %q", got.Resources[0].Name, "starter")
		}
		if got.Resources[0].Latest != "1.0.0" {
			t.Errorf("resource latest: got %q, want %q", got.Resources[0].Latest, "1.0.0")
		}
		if !got.Resources[0].UpdatedAt.Equal(time.Unix(1000000, 0)) {
			t.Errorf("resource updated_at: got %v, want %v", got.Resources[0].UpdatedAt, time.Unix(1000000, 0))
		}
	})
}

func TestPutNamespace(t *testing.T) {
	c, err := openCacheWithPath(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening cache: %v", err)
	}
	defer c.Close()

	t.Run("insert", func(t *testing.T) {
		info := &NamespaceInfo{
			Namespace:   "crucible",
			Description: "Official templates",
			Resources: []ResourceSummary{
				{
					Name:        "starter",
					Type:        "template",
					Description: "A starter template",
					Latest:      "1.0.0",
					UpdatedAt:   time.Now(),
				},
				{
					Name:        "advanced",
					Type:        "template",
					Description: "An advanced template",
					Latest:      "2.0.0",
					UpdatedAt:   time.Now(),
				},
			},
		}

		err := c.PutNamespace(info, "etag-123")
		if err != nil {
			t.Fatalf("putting namespace: %v", err)
		}

		got, _, err := c.GetNamespace("crucible")
		if err != nil {
			t.Fatalf("getting namespace: %v", err)
		}

		if len(got.Resources) != 2 {
			t.Errorf("resources: got %d, want 2", len(got.Resources))
		}
	})

	t.Run("replace", func(t *testing.T) {
		updated := &NamespaceInfo{
			Namespace:   "crucible",
			Description: "Updated description",
			Resources: []ResourceSummary{
				{
					Name:        "new-resource",
					Type:        "template",
					Description: "A new resource",
					Latest:      "3.0.0",
					UpdatedAt:   time.Now(),
				},
			},
		}

		err := c.PutNamespace(updated, "etag-456")
		if err != nil {
			t.Fatalf("putting namespace: %v", err)
		}

		got, etag, err := c.GetNamespace("crucible")
		if err != nil {
			t.Fatalf("getting namespace: %v", err)
		}

		if etag != "etag-456" {
			t.Errorf("etag: got %q, want %q", etag, "etag-456")
		}
		if got.Description != "Updated description" {
			t.Errorf("description: got %q, want %q", got.Description, "Updated description")
		}
		if len(got.Resources) != 1 {
			t.Fatalf("resources: got %d, want 1", len(got.Resources))
		}
		if got.Resources[0].Name != "new-resource" {
			t.Errorf("resource name: got %q, want %q", got.Resources[0].Name, "new-resource")
		}
	})

	t.Run("empty resources", func(t *testing.T) {
		info := &NamespaceInfo{
			Namespace:   "empty",
			Description: "Empty namespace",
			Resources:   nil,
		}

		err := c.PutNamespace(info, "etag-789")
		if err != nil {
			t.Fatalf("putting namespace: %v", err)
		}

		got, _, err := c.GetNamespace("empty")
		if err != nil {
			t.Fatalf("getting namespace: %v", err)
		}

		if len(got.Resources) != 0 {
			t.Errorf("resources: got %d, want 0", len(got.Resources))
		}
	})
}
