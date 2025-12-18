package store

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/cruciblehq/crux/pkg/reference"
)

func TestGetResource(t *testing.T) {
	c, err := openCacheWithPath(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening cache: %v", err)
	}
	defer c.Close()

	t.Run("not found", func(t *testing.T) {
		_, _, err := c.GetResource("crucible", "starter")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("found", func(t *testing.T) {
		version, _ := reference.ParseVersion("1.0.0")
		digest, _ := reference.ParseDigest("sha256:abc123")

		info := &ResourceInfo{
			Name:        "starter",
			Type:        "template",
			Description: "A starter template",
			Versions: []VersionInfo{
				{
					Version:   *version,
					Digest:    *digest,
					Published: time.Unix(1000000, 0),
					Size:      1024,
				},
			},
			Channels: []ChannelInfo{
				{
					VersionInfo: VersionInfo{
						Version:   *version,
						Digest:    *digest,
						Published: time.Unix(1000000, 0),
						Size:      1024,
					},
					Channel:     "stable",
					Description: "Stable release channel",
				},
			},
		}

		err := c.PutResource("crucible", info, "etag-123")
		if err != nil {
			t.Fatalf("putting resource: %v", err)
		}

		got, etag, err := c.GetResource("crucible", "starter")
		if err != nil {
			t.Fatalf("getting resource: %v", err)
		}

		if etag != "etag-123" {
			t.Errorf("etag: got %q, want %q", etag, "etag-123")
		}
		if got.Name != info.Name {
			t.Errorf("name: got %q, want %q", got.Name, info.Name)
		}
		if got.Type != info.Type {
			t.Errorf("type: got %q, want %q", got.Type, info.Type)
		}
		if got.Description != info.Description {
			t.Errorf("description: got %q, want %q", got.Description, info.Description)
		}
		if len(got.Versions) != 1 {
			t.Fatalf("versions: got %d, want 1", len(got.Versions))
		}
		if got.Versions[0].Version.String() != "1.0.0" {
			t.Errorf("version: got %q, want %q", got.Versions[0].Version.String(), "1.0.0")
		}
		if got.Versions[0].Digest.String() != "sha256:abc123" {
			t.Errorf("digest: got %q, want %q", got.Versions[0].Digest.String(), "sha256:abc123")
		}
		if len(got.Channels) != 1 {
			t.Fatalf("channels: got %d, want 1", len(got.Channels))
		}
		if got.Channels[0].Channel != "stable" {
			t.Errorf("channel: got %q, want %q", got.Channels[0].Channel, "stable")
		}
	})
}

func TestPutResource(t *testing.T) {
	c, err := openCacheWithPath(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening cache: %v", err)
	}
	defer c.Close()

	version1, _ := reference.ParseVersion("1.0.0")
	version2, _ := reference.ParseVersion("2.0.0")
	digest1, _ := reference.ParseDigest("sha256:abc123")
	digest2, _ := reference.ParseDigest("sha256:def456")

	t.Run("insert", func(t *testing.T) {
		info := &ResourceInfo{
			Name:        "starter",
			Type:        "template",
			Description: "A starter template",
			Versions: []VersionInfo{
				{
					Version:   *version1,
					Digest:    *digest1,
					Published: time.Now(),
					Size:      1024,
				},
				{
					Version:   *version2,
					Digest:    *digest2,
					Published: time.Now(),
					Size:      2048,
				},
			},
			Channels: []ChannelInfo{
				{
					VersionInfo: VersionInfo{
						Version:   *version2,
						Digest:    *digest2,
						Published: time.Now(),
						Size:      2048,
					},
					Channel:     "stable",
					Description: "Stable release channel",
				},
			},
		}

		err := c.PutResource("crucible", info, "etag-123")
		if err != nil {
			t.Fatalf("putting resource: %v", err)
		}

		got, _, err := c.GetResource("crucible", "starter")
		if err != nil {
			t.Fatalf("getting resource: %v", err)
		}

		if len(got.Versions) != 2 {
			t.Errorf("versions: got %d, want 2", len(got.Versions))
		}
		if len(got.Channels) != 1 {
			t.Errorf("channels: got %d, want 1", len(got.Channels))
		}
	})

	t.Run("replace", func(t *testing.T) {
		updated := &ResourceInfo{
			Name:        "starter",
			Type:        "template",
			Description: "Updated description",
			Versions: []VersionInfo{
				{
					Version:   *version1,
					Digest:    *digest1,
					Published: time.Now(),
					Size:      1024,
				},
			},
			Channels: nil,
		}

		err := c.PutResource("crucible", updated, "etag-456")
		if err != nil {
			t.Fatalf("putting resource: %v", err)
		}

		got, etag, err := c.GetResource("crucible", "starter")
		if err != nil {
			t.Fatalf("getting resource: %v", err)
		}

		if etag != "etag-456" {
			t.Errorf("etag: got %q, want %q", etag, "etag-456")
		}
		if got.Description != "Updated description" {
			t.Errorf("description: got %q, want %q", got.Description, "Updated description")
		}
		if len(got.Versions) != 1 {
			t.Errorf("versions: got %d, want 1", len(got.Versions))
		}
		if len(got.Channels) != 0 {
			t.Errorf("channels: got %d, want 0", len(got.Channels))
		}
	})

	t.Run("empty versions and channels", func(t *testing.T) {
		info := &ResourceInfo{
			Name:        "empty",
			Type:        "template",
			Description: "Empty resource",
			Versions:    nil,
			Channels:    nil,
		}

		err := c.PutResource("crucible", info, "etag-789")
		if err != nil {
			t.Fatalf("putting resource: %v", err)
		}

		got, _, err := c.GetResource("crucible", "empty")
		if err != nil {
			t.Fatalf("getting resource: %v", err)
		}

		if len(got.Versions) != 0 {
			t.Errorf("versions: got %d, want 0", len(got.Versions))
		}
		if len(got.Channels) != 0 {
			t.Errorf("channels: got %d, want 0", len(got.Channels))
		}
	})
}
