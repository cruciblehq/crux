package store

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/cruciblehq/crux/pkg/reference"
)

func newTestRemote(t *testing.T, server *httptest.Server) *remoteImpl {
	t.Helper()

	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse server URL: %v", err)
	}

	return newRemoteWithHTTPClient(server.Client(), baseURL, RemoteOptions{})
}

func mustParseIdentifier(t *testing.T, s string) reference.Identifier {
	t.Helper()

	id, err := reference.ParseIdentifier(s, "resource", nil)
	if err != nil {
		t.Fatalf("failed to parse identifier %q: %v", s, err)
	}

	return *id
}

func TestNamespace(t *testing.T) {
	info := NamespaceInfo{
		Resources: []ResourceSummary{
			{Name: "foo"},
			{Name: "bar"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/test-namespace" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get(headerAccept) != MediaTypeNamespace {
			t.Errorf("unexpected Accept header: %s", r.Header.Get(headerAccept))
		}

		w.Header().Set(headerContentType, MediaTypeNamespace)
		w.Header().Set(headerETag, `"abc123"`)
		json.NewEncoder(w).Encode(info)
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	resp, err := remote.Namespace(context.Background(), &NamespaceRequest{
		Namespace: "test-namespace",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ETag != `"abc123"` {
		t.Errorf("expected ETag %q, got %q", `"abc123"`, resp.ETag)
	}
	if resp.Info == nil {
		t.Fatal("expected Info to be set")
	}
	if len(resp.Info.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resp.Info.Resources))
	}
}

func TestNamespaceNotModified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(headerIfNoneMatch) != `"abc123"` {
			t.Errorf("expected If-None-Match header, got %q", r.Header.Get(headerIfNoneMatch))
		}

		w.Header().Set(headerETag, `"abc123"`)
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	resp, err := remote.Namespace(context.Background(), &NamespaceRequest{
		Namespace: "test-namespace",
		ETag:      `"abc123"`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ETag != `"abc123"` {
		t.Errorf("expected ETag %q, got %q", `"abc123"`, resp.ETag)
	}
	if resp.Info != nil {
		t.Error("expected Info to be nil for 304 response")
	}
}

func TestNamespaceNotModifiedPreservesETag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	resp, err := remote.Namespace(context.Background(), &NamespaceRequest{
		Namespace: "test-namespace",
		ETag:      `"original"`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ETag != `"original"` {
		t.Errorf("expected preserved ETag %q, got %q", `"original"`, resp.ETag)
	}
}

func TestResource(t *testing.T) {
	info := ResourceInfo{
		Name: "test-resource",
		Type: "template",
		Versions: []VersionInfo{
			{Version: *mustParseVersion(t, "1.0.0")},
			{Version: *mustParseVersion(t, "2.0.0")},
		},
		Channels: []ChannelInfo{
			{VersionInfo: VersionInfo{Version: *mustParseVersion(t, "2.0.0")}, Channel: "latest"},
			{VersionInfo: VersionInfo{Version: *mustParseVersion(t, "1.0.0")}, Channel: "stable"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/test-namespace/test-resource" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get(headerAccept) != MediaTypeResource {
			t.Errorf("unexpected Accept header: %s", r.Header.Get(headerAccept))
		}

		w.Header().Set(headerContentType, MediaTypeResource)
		w.Header().Set(headerETag, `"def456"`)
		json.NewEncoder(w).Encode(info)
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	resp, err := remote.Resource(context.Background(), &ResourceRequest{
		Identifier: mustParseIdentifier(t, "test-namespace/test-resource"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ETag != `"def456"` {
		t.Errorf("expected ETag %q, got %q", `"def456"`, resp.ETag)
	}
	if resp.Info == nil {
		t.Fatal("expected Info to be set")
	}
	if len(resp.Info.Versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(resp.Info.Versions))
	}
	if len(resp.Info.Channels) != 2 {
		t.Errorf("expected 2 channels, got %d", len(resp.Info.Channels))
	}
}

func TestResourceNotModified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerETag, `"def456"`)
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	resp, err := remote.Resource(context.Background(), &ResourceRequest{
		Identifier: mustParseIdentifier(t, "test-namespace/test-resource"),
		ETag:       `"def456"`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Info != nil {
		t.Error("expected Info to be nil for 304 response")
	}
}

func TestFetch(t *testing.T) {
	archiveData := []byte("fake archive data")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/test-namespace/test-resource/1.2.3" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get(headerAccept) != MediaTypeArchive {
			t.Errorf("unexpected Accept header: %s", r.Header.Get(headerAccept))
		}

		w.Header().Set(headerContentType, MediaTypeArchive)
		w.Header().Set(headerETag, `"archive123"`)
		w.Write(archiveData)
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	resp, err := remote.Fetch(context.Background(), &FetchRequest{
		Identifier: mustParseIdentifier(t, "test-namespace/test-resource"),
		Version:    *mustParseVersion(t, "1.2.3"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ETag != `"archive123"` {
		t.Errorf("expected ETag %q, got %q", `"archive123"`, resp.ETag)
	}
	if string(resp.Data) != string(archiveData) {
		t.Errorf("expected data %q, got %q", archiveData, resp.Data)
	}
}

func TestFetchNotModified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerETag, `"archive123"`)
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	resp, err := remote.Fetch(context.Background(), &FetchRequest{
		Identifier: mustParseIdentifier(t, "test-namespace/test-resource"),
		Version:    *mustParseVersion(t, "1.2.3"),
		ETag:       `"archive123"`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Data != nil {
		t.Error("expected Data to be nil for 304 response")
	}
}

func TestConsume(t *testing.T) {
	archiveData := []byte("fake archive data")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/test-namespace/test-resource/:stable" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set(headerContentType, MediaTypeArchive)
		w.Header().Set(headerETag, `"consume123"`)
		w.Header().Set(headerContentVersion, "2.1.0")
		w.Write(archiveData)
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	resp, err := remote.Consume(context.Background(), &ConsumeRequest{
		Identifier: mustParseIdentifier(t, "test-namespace/test-resource"),
		Channel:    "stable",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ETag != `"consume123"` {
		t.Errorf("expected ETag %q, got %q", `"consume123"`, resp.ETag)
	}
	if resp.Version.String() != "2.1.0" {
		t.Errorf("expected version 2.1.0, got %s", resp.Version.String())
	}
	if string(resp.Data) != string(archiveData) {
		t.Errorf("expected data %q, got %q", archiveData, resp.Data)
	}
}

func TestConsumeNotModified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerETag, `"consume123"`)
		w.Header().Set(headerContentVersion, "2.1.0")
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	resp, err := remote.Consume(context.Background(), &ConsumeRequest{
		Identifier: mustParseIdentifier(t, "test-namespace/test-resource"),
		Channel:    "stable",
		ETag:       `"consume123"`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Data != nil {
		t.Error("expected Data to be nil for 304 response")
	}
	if resp.Version.String() != "2.1.0" {
		t.Errorf("expected version 2.1.0, got %s", resp.Version.String())
	}
}

func TestConsumeMissingVersionHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerContentType, MediaTypeArchive)
		w.Write([]byte("data"))
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	_, err := remote.Consume(context.Background(), &ConsumeRequest{
		Identifier: mustParseIdentifier(t, "test-namespace/test-resource"),
		Channel:    "stable",
	})
	if err == nil {
		t.Fatal("expected error for missing Content-Version header")
	}
}

func TestRegistryError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerContentType, MediaTypeError)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(RegistryError{
			Code:    "not_found",
			Message: "resource not found",
		})
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	_, err := remote.Namespace(context.Background(), &NamespaceRequest{
		Namespace: "missing",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	regErr, ok := err.(*RegistryError)
	if !ok {
		t.Fatalf("expected *RegistryError, got %T", err)
	}
	if regErr.Code != "not_found" {
		t.Errorf("expected code 'not_found', got %q", regErr.Code)
	}
}

func TestInvalidContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerContentType, "text/plain")
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	_, err := remote.Namespace(context.Background(), &NamespaceRequest{
		Namespace: "test",
	})
	if err == nil {
		t.Fatal("expected error for invalid content type")
	}
}

func TestMissingContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	_, err := remote.Namespace(context.Background(), &NamespaceRequest{
		Namespace: "test",
	})
	if err == nil {
		t.Fatal("expected error for missing content type")
	}
}

func TestArchiveTooLarge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerContentType, MediaTypeArchive)
		w.Header().Set(headerContentLength, "200000000")
		w.Write([]byte("data"))
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	_, err := remote.Fetch(context.Background(), &FetchRequest{
		Identifier: mustParseIdentifier(t, "test-namespace/test-resource"),
		Version:    *mustParseVersion(t, "1.0.0"),
	})
	if err == nil {
		t.Fatal("expected error for archive too large")
	}
}

func TestUserAgent(t *testing.T) {
	var receivedUserAgent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserAgent = r.Header.Get(headerUserAgent)
		w.Header().Set(headerContentType, MediaTypeNamespace)
		json.NewEncoder(w).Encode(NamespaceInfo{})
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	remote := newRemoteWithHTTPClient(server.Client(), baseURL, RemoteOptions{
		UserAgent: "Crucible/1.2.3",
	})

	_, err := remote.Namespace(context.Background(), &NamespaceRequest{
		Namespace: "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedUserAgent != "Crucible/1.2.3" {
		t.Errorf("expected User-Agent 'Crucible/1.2.3', got %q", receivedUserAgent)
	}
}

func TestMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerContentType, MediaTypeNamespace)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	_, err := remote.Namespace(context.Background(), &NamespaceRequest{
		Namespace: "test-namespace",
	})
	if err == nil {
		t.Fatal("expected error for malformed response")
	}
}

func TestServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	remote := newTestRemote(t, server)

	_, err := remote.Namespace(context.Background(), &NamespaceRequest{
		Namespace: "test-namespace",
	})
	if err == nil {
		t.Fatal("expected error for 500 Internal Server Error")
	}
}
