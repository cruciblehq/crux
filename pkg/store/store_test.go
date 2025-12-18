package store

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cruciblehq/crux/pkg/paths"
	"github.com/cruciblehq/crux/pkg/reference"
	"github.com/klauspost/compress/zstd"
)

type mockRemote struct {
	namespaceFunc func(ctx context.Context, req *NamespaceRequest) (*NamespaceResponse, error)
	resourceFunc  func(ctx context.Context, req *ResourceRequest) (*ResourceResponse, error)
	fetchFunc     func(ctx context.Context, req *FetchRequest) (*FetchResponse, error)
	consumeFunc   func(ctx context.Context, req *ConsumeRequest) (*ConsumeResponse, error)
}

func (m *mockRemote) Namespace(ctx context.Context, req *NamespaceRequest) (*NamespaceResponse, error) {
	if m.namespaceFunc != nil {
		return m.namespaceFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRemote) Resource(ctx context.Context, req *ResourceRequest) (*ResourceResponse, error) {
	if m.resourceFunc != nil {
		return m.resourceFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRemote) Fetch(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRemote) Consume(ctx context.Context, req *ConsumeRequest) (*ConsumeResponse, error) {
	if m.consumeFunc != nil {
		return m.consumeFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

func createTestArchive(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer

	zw, err := zstd.NewWriter(&buf)
	if err != nil {
		t.Fatalf("creating zstd writer: %v", err)
	}

	tw := tar.NewWriter(zw)

	content := []byte("test file content")
	hdr := &tar.Header{
		Name: "test.txt",
		Mode: int64(paths.DefaultFileMode),
		Size: int64(len(content)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("writing tar header: %v", err)
	}

	if _, err := tw.Write(content); err != nil {
		t.Fatalf("writing tar content: %v", err)
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("closing tar writer: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("closing zstd writer: %v", err)
	}

	return buf.Bytes()
}

func TestNew(t *testing.T) {
	cache, err := openCacheWithPath(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening cache: %v", err)
	}
	defer cache.Close()

	remote := &mockRemote{}

	t.Run("success", func(t *testing.T) {
		s, err := New(cache, remote)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s == nil {
			t.Fatal("expected store, got nil")
		}
	})

	t.Run("nil cache", func(t *testing.T) {
		_, err := New(nil, remote)
		if !errors.Is(err, ErrCacheRequired) {
			t.Errorf("expected ErrCacheRequired, got %v", err)
		}
	})

	t.Run("nil remote", func(t *testing.T) {
		_, err := New(cache, nil)
		if !errors.Is(err, ErrRemoteRequired) {
			t.Errorf("expected ErrRemoteRequired, got %v", err)
		}
	})
}

func TestResolve(t *testing.T) {
	archiveData := createTestArchive(t)
	hash := sha256.Sum256(archiveData)
	digest := mustParseDigest(t, "sha256:"+hex.EncodeToString(hash[:]))
	version := mustParseVersion(t, "1.0.0")

	resourceInfo := &ResourceInfo{
		Name: "starter",
		Type: "template",
		Versions: []VersionInfo{
			{
				Version:   *version,
				Digest:    *digest,
				Published: time.Now(),
				Size:      int64(len(archiveData)),
			},
		},
	}

	t.Run("fetches and caches", func(t *testing.T) {
		tmpDir := t.TempDir()
		cache, err := openCacheWithPath(filepath.Join(tmpDir, "test.db"))
		if err != nil {
			t.Fatalf("opening cache: %v", err)
		}
		defer cache.Close()

		fetchCalled := false
		remote := &mockRemote{
			resourceFunc: func(ctx context.Context, req *ResourceRequest) (*ResourceResponse, error) {
				return &ResourceResponse{
					Info: resourceInfo,
					ETag: "etag-123",
				}, nil
			},
			fetchFunc: func(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
				fetchCalled = true
				return &FetchResponse{
					Data: archiveData,
					ETag: "archive-etag",
				}, nil
			},
		}

		s, err := newWithPath(cache, remote, filepath.Join(tmpDir, "store"))
		if err != nil {
			t.Fatalf("creating store: %v", err)
		}

		ref := mustParseRef(t, "crucible/starter 1.0.0")

		path, err := s.Resolve(context.Background(), ref)
		if err != nil {
			t.Fatalf("resolving: %v", err)
		}

		if !fetchCalled {
			t.Error("expected fetch to be called")
		}

		if path == "" {
			t.Error("expected non-empty path")
		}
	})

	t.Run("uses cache on 304", func(t *testing.T) {
		tmpDir := t.TempDir()
		cache, err := openCacheWithPath(filepath.Join(tmpDir, "test.db"))
		if err != nil {
			t.Fatalf("opening cache: %v", err)
		}
		defer cache.Close()

		err = cache.PutResource("crucible", resourceInfo, "etag-123")
		if err != nil {
			t.Fatalf("putting resource: %v", err)
		}

		resourceCalls := 0
		remote := &mockRemote{
			resourceFunc: func(ctx context.Context, req *ResourceRequest) (*ResourceResponse, error) {
				resourceCalls++
				if req.ETag == "etag-123" {
					return &ResourceResponse{
						Info: nil,
						ETag: "etag-123",
					}, nil
				}
				return &ResourceResponse{
					Info: resourceInfo,
					ETag: "etag-123",
				}, nil
			},
			fetchFunc: func(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
				return &FetchResponse{
					Data: archiveData,
					ETag: "archive-etag",
				}, nil
			},
		}

		s, err := newWithPath(cache, remote, filepath.Join(tmpDir, "store"))
		if err != nil {
			t.Fatalf("creating store: %v", err)
		}

		ref := mustParseRef(t, "crucible/starter 1.0.0")

		_, err = s.Resolve(context.Background(), ref)
		if err != nil {
			t.Fatalf("resolving: %v", err)
		}

		if resourceCalls != 1 {
			t.Errorf("expected 1 resource call, got %d", resourceCalls)
		}
	})

	t.Run("uses cached archive", func(t *testing.T) {
		tmpDir := t.TempDir()
		cache, err := openCacheWithPath(filepath.Join(tmpDir, "test.db"))
		if err != nil {
			t.Fatalf("opening cache: %v", err)
		}
		defer cache.Close()

		err = cache.PutResource("crucible", resourceInfo, "etag-123")
		if err != nil {
			t.Fatalf("putting resource: %v", err)
		}

		extractPath := filepath.Join(tmpDir, "store", "crucible", "starter", "1.0.0")
		err = os.MkdirAll(extractPath, 0755)
		if err != nil {
			t.Fatalf("creating extract path: %v", err)
		}

		err = cache.PutArchive(&Archive{
			Namespace: "crucible",
			Name:      "starter",
			Version:   *version,
			Digest:    *digest,
			Path:      extractPath,
			ETag:      "archive-etag",
		})
		if err != nil {
			t.Fatalf("putting archive: %v", err)
		}

		fetchCalled := false
		remote := &mockRemote{
			resourceFunc: func(ctx context.Context, req *ResourceRequest) (*ResourceResponse, error) {
				if req.ETag == "etag-123" {
					return &ResourceResponse{Info: nil, ETag: "etag-123"}, nil
				}
				return &ResourceResponse{Info: resourceInfo, ETag: "etag-123"}, nil
			},
			fetchFunc: func(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
				fetchCalled = true
				return &FetchResponse{Data: archiveData, ETag: "archive-etag"}, nil
			},
		}

		s, err := newWithPath(cache, remote, filepath.Join(tmpDir, "store"))
		if err != nil {
			t.Fatalf("creating store: %v", err)
		}

		ref := mustParseRef(t, "crucible/starter 1.0.0")

		path, err := s.Resolve(context.Background(), ref)
		if err != nil {
			t.Fatalf("resolving: %v", err)
		}

		if fetchCalled {
			t.Error("expected fetch not to be called when archive is cached")
		}

		if path != extractPath {
			t.Errorf("expected path %s, got %s", extractPath, path)
		}
	})

	t.Run("removes stale archive cache entry", func(t *testing.T) {
		tmpDir := t.TempDir()
		cache, err := openCacheWithPath(filepath.Join(tmpDir, "test.db"))
		if err != nil {
			t.Fatalf("opening cache: %v", err)
		}
		defer cache.Close()

		err = cache.PutResource("crucible", resourceInfo, "etag-123")
		if err != nil {
			t.Fatalf("putting resource: %v", err)
		}

		missingPath := filepath.Join(tmpDir, "store", "crucible", "starter", "1.0.0")
		err = cache.PutArchive(&Archive{
			Namespace: "crucible",
			Name:      "starter",
			Version:   *version,
			Digest:    *digest,
			Path:      missingPath,
			ETag:      "archive-etag",
		})
		if err != nil {
			t.Fatalf("putting archive: %v", err)
		}

		fetchCalled := false
		remote := &mockRemote{
			resourceFunc: func(ctx context.Context, req *ResourceRequest) (*ResourceResponse, error) {
				if req.ETag == "etag-123" {
					return &ResourceResponse{Info: nil, ETag: "etag-123"}, nil
				}
				return &ResourceResponse{Info: resourceInfo, ETag: "etag-123"}, nil
			},
			fetchFunc: func(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
				fetchCalled = true
				return &FetchResponse{Data: archiveData, ETag: "archive-etag-new"}, nil
			},
		}

		s, err := newWithPath(cache, remote, filepath.Join(tmpDir, "store"))
		if err != nil {
			t.Fatalf("creating store: %v", err)
		}

		ref := mustParseRef(t, "crucible/starter 1.0.0")

		_, err = s.Resolve(context.Background(), ref)
		if err != nil {
			t.Fatalf("resolving: %v", err)
		}

		if !fetchCalled {
			t.Error("expected fetch to be called when archive path is missing")
		}
	})
}

func TestSelectVersion(t *testing.T) {
	v100 := mustParseVersion(t, "1.0.0")
	v110 := mustParseVersion(t, "1.1.0")
	v200 := mustParseVersion(t, "2.0.0")
	digest := mustParseDigest(t, "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

	info := &ResourceInfo{
		Name: "test",
		Versions: []VersionInfo{
			{Version: *v100, Digest: *digest},
			{Version: *v110, Digest: *digest},
			{Version: *v200, Digest: *digest},
		},
		Channels: []ChannelInfo{
			{VersionInfo: VersionInfo{Version: *v110, Digest: *digest}, Channel: "stable"},
			{VersionInfo: VersionInfo{Version: *v200, Digest: *digest}, Channel: "latest"},
		},
	}

	cache, err := openCacheWithPath(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening cache: %v", err)
	}
	defer cache.Close()

	s, _ := newWithPath(cache, &mockRemote{}, t.TempDir())

	t.Run("exact version", func(t *testing.T) {
		ref := mustParseRef(t, "test/test 1.0.0")
		v, err := s.selectVersion(ref, info)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.String() != "1.0.0" {
			t.Errorf("expected 1.0.0, got %s", v.String())
		}
	})

	t.Run("caret constraint", func(t *testing.T) {
		ref := mustParseRef(t, "test/test ^1.0.0")
		v, err := s.selectVersion(ref, info)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.String() != "1.1.0" {
			t.Errorf("expected 1.1.0, got %s", v.String())
		}
	})

	t.Run("channel", func(t *testing.T) {
		ref := mustParseRef(t, "test/test :stable")
		v, err := s.selectVersion(ref, info)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.String() != "1.1.0" {
			t.Errorf("expected 1.1.0, got %s", v.String())
		}
	})

	t.Run("unknown channel", func(t *testing.T) {
		ref := mustParseRef(t, "test/test :beta")
		_, err := s.selectVersion(ref, info)
		if !errors.Is(err, ErrNoMatchingVersion) {
			t.Errorf("expected ErrNoMatchingVersion, got %v", err)
		}
	})

	t.Run("no matching version", func(t *testing.T) {
		ref := mustParseRef(t, "test/test ^3.0.0")
		_, err := s.selectVersion(ref, info)
		if !errors.Is(err, ErrNoMatchingVersion) {
			t.Errorf("expected ErrNoMatchingVersion, got %v", err)
		}
	})
}

func TestVerifyDigest(t *testing.T) {
	data := []byte("test data")
	hash := sha256.Sum256(data)
	validDigest := mustParseDigest(t, "sha256:"+hex.EncodeToString(hash[:]))
	invalidDigest := mustParseDigest(t, "sha256:0000000000000000000000000000000000000000000000000000000000000000")

	t.Run("valid", func(t *testing.T) {
		err := verifyDigest(data, validDigest)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		err := verifyDigest(data, invalidDigest)
		if !errors.Is(err, ErrDigestMismatch) {
			t.Errorf("expected ErrDigestMismatch, got %v", err)
		}
	})
}

func TestFindDigest(t *testing.T) {
	v100 := mustParseVersion(t, "1.0.0")
	v200 := mustParseVersion(t, "2.0.0")
	v300 := mustParseVersion(t, "3.0.0")
	digest := mustParseDigest(t, "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

	info := &ResourceInfo{
		Name: "test",
		Versions: []VersionInfo{
			{Version: *v100, Digest: *digest},
			{Version: *v200, Digest: *digest},
		},
	}

	t.Run("found", func(t *testing.T) {
		d, err := findDigest(v100, info)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d.String() != digest.String() {
			t.Errorf("expected %s, got %s", digest.String(), d.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := findDigest(v300, info)
		if !errors.Is(err, ErrNoMatchingVersion) {
			t.Errorf("expected ErrNoMatchingVersion, got %v", err)
		}
	})
}

func mustParseRef(t *testing.T, s string) *reference.Reference {
	t.Helper()
	ref, err := reference.Parse(s, "template", nil)
	if err != nil {
		t.Fatalf("parsing reference %q: %v", s, err)
	}
	return ref
}

func mustParseVersion(t *testing.T, s string) *reference.Version {
	t.Helper()
	v, err := reference.ParseVersion(s)
	if err != nil {
		t.Fatalf("parsing version %q: %v", s, err)
	}
	return v
}

func mustParseDigest(t *testing.T, s string) *reference.Digest {
	t.Helper()
	d, err := reference.ParseDigest(s)
	if err != nil {
		t.Fatalf("parsing digest %q: %v", s, err)
	}
	return d
}
