package registry

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_CreateNamespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces" {
			t.Errorf("expected /namespaces, got %s", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "namespace-info") {
			t.Errorf("expected namespace-info content type, got %s", ct)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.namespace.v0+json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"name":"test","description":"","resources":[],"createdAt":0,"updatedAt":0}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	ns, err := client.CreateNamespace(context.Background(), NamespaceInfo{
		Name:        "test",
		Description: "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ns.Name != "test" {
		t.Errorf("expected name 'test', got %s", ns.Name)
	}
}

func TestClient_ReadNamespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test" {
			t.Errorf("expected /namespaces/test, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.namespace.v0+json")
		w.Write([]byte(`{"name":"test","description":"","resources":[],"createdAt":0,"updatedAt":0}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	ns, err := client.ReadNamespace(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ns.Name != "test" {
		t.Errorf("expected name 'test', got %s", ns.Name)
	}
}

func TestClient_UpdateNamespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test" {
			t.Errorf("expected /namespaces/test, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.namespace.v0+json")
		w.Write([]byte(`{"name":"test","description":"updated","resources":[],"createdAt":0,"updatedAt":1}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	ns, err := client.UpdateNamespace(context.Background(), "test", NamespaceInfo{
		Name:        "test",
		Description: "updated",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ns.Description != "updated" {
		t.Errorf("expected description 'updated', got %s", ns.Description)
	}
}

func TestClient_DeleteNamespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test" {
			t.Errorf("expected /namespaces/test, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	err := client.DeleteNamespace(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ListNamespaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces" {
			t.Errorf("expected /namespaces, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.namespace-list.v0+json")
		w.Write([]byte(`{"namespaces":[{"name":"test","description":"","resourceCount":0,"createdAt":0,"updatedAt":0}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	list, err := client.ListNamespaces(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Namespaces) != 1 {
		t.Errorf("expected 1 namespace, got %d", len(list.Namespaces))
	}
}

func TestClient_CreateResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources" {
			t.Errorf("expected /namespaces/test/resources, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.resource.v0+json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"namespace":"test","name":"myres","type":"widget","description":"","versions":[],"channels":[],"createdAt":0,"updatedAt":0}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	res, err := client.CreateResource(context.Background(), "test", ResourceInfo{
		Name:        "myres",
		Type:        "widget",
		Description: "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Name != "myres" {
		t.Errorf("expected name 'myres', got %s", res.Name)
	}
}

func TestClient_ReadResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres" {
			t.Errorf("expected /namespaces/test/resources/myres, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.resource.v0+json")
		w.Write([]byte(`{"namespace":"test","name":"myres","type":"widget","description":"","versions":[],"channels":[],"createdAt":0,"updatedAt":0}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	res, err := client.ReadResource(context.Background(), "test", "myres")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Name != "myres" {
		t.Errorf("expected name 'myres', got %s", res.Name)
	}
}

func TestClient_UpdateResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres" {
			t.Errorf("expected /namespaces/test/resources/myres, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.resource.v0+json")
		w.Write([]byte(`{"namespace":"test","name":"myres","type":"widget","description":"updated","versions":[],"channels":[],"createdAt":0,"updatedAt":1}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	res, err := client.UpdateResource(context.Background(), "test", "myres", ResourceInfo{
		Name:        "myres",
		Type:        "widget",
		Description: "updated",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Description != "updated" {
		t.Errorf("expected description 'updated', got %s", res.Description)
	}
}

func TestClient_DeleteResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres" {
			t.Errorf("expected /namespaces/test/resources/myres, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	err := client.DeleteResource(context.Background(), "test", "myres")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ListResources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources" {
			t.Errorf("expected /namespaces/test/resources, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.resource-list.v0+json")
		w.Write([]byte(`{"resources":[{"namespace":"test","name":"myres","type":"widget","description":"","versionCount":0,"channelCount":0,"createdAt":0,"updatedAt":0}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	list, err := client.ListResources(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(list.Resources))
	}
}

func TestClient_CreateVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres/versions" {
			t.Errorf("expected /namespaces/test/resources/myres/versions, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.version.v0+json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"namespace":"test","resource":"myres","string":"1.0.0","archive":null,"size":null,"digest":null,"createdAt":0,"updatedAt":0}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	ver, err := client.CreateVersion(context.Background(), "test", "myres", VersionInfo{
		String: "1.0.0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ver.String != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", ver.String)
	}
}

func TestClient_ReadVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres/versions/1.0.0" {
			t.Errorf("expected /namespaces/test/resources/myres/versions/1.0.0, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.version.v0+json")
		w.Write([]byte(`{"namespace":"test","resource":"myres","string":"1.0.0","archive":null,"size":null,"digest":null,"createdAt":0,"updatedAt":0}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	ver, err := client.ReadVersion(context.Background(), "test", "myres", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ver.String != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", ver.String)
	}
}

func TestClient_UpdateVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres/versions/1.0.0" {
			t.Errorf("expected /namespaces/test/resources/myres/versions/1.0.0, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.version.v0+json")
		w.Write([]byte(`{"namespace":"test","resource":"myres","string":"1.0.0","archive":null,"size":null,"digest":null,"createdAt":0,"updatedAt":1}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	ver, err := client.UpdateVersion(context.Background(), "test", "myres", "1.0.0", VersionInfo{
		String: "1.0.0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ver.UpdatedAt != 1 {
		t.Errorf("expected updatedAt '1', got %d", ver.UpdatedAt)
	}
}

func TestClient_DeleteVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres/versions/1.0.0" {
			t.Errorf("expected /namespaces/test/resources/myres/versions/1.0.0, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	err := client.DeleteVersion(context.Background(), "test", "myres", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ListVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres/versions" {
			t.Errorf("expected /namespaces/test/resources/myres/versions, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.version-list.v0+json")
		w.Write([]byte(`{"versions":[{"namespace":"test","resource":"myres","string":"1.0.0","description":"","hasArchive":true,"createdAt":0,"updatedAt":0}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	list, err := client.ListVersions(context.Background(), "test", "myres")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Versions) != 1 {
		t.Errorf("expected 1 version, got %d", len(list.Versions))
	}
}

func TestClient_UploadArchive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres/versions/1.0.0/archive" {
			t.Errorf("expected /namespaces/test/resources/myres/versions/1.0.0/archive, got %s", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != string(MediaTypeArchive) {
			t.Errorf("expected archive content type, got %s", ct)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.version.v0+json")
		w.Write([]byte(`{"namespace":"test","resource":"myres","string":"1.0.0","archive":"url","size":1024,"digest":"abc123","createdAt":0,"updatedAt":1}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	archive := bytes.NewReader([]byte("fake archive data"))
	ver, err := client.UploadArchive(context.Background(), "test", "myres", "1.0.0", archive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ver.Archive == nil || *ver.Archive != "url" {
		t.Errorf("expected archive URL 'url'")
	}
}

func TestClient_DownloadArchive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres/versions/1.0.0/archive" {
			t.Errorf("expected /namespaces/test/resources/myres/versions/1.0.0/archive, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", string(MediaTypeArchive))
		w.Write([]byte("fake archive data"))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	rc, err := client.DownloadArchive(context.Background(), "test", "myres", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("failed to read archive: %v", err)
	}
	if string(data) != "fake archive data" {
		t.Errorf("expected 'fake archive data', got %s", string(data))
	}
}

func TestClient_CreateChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres/channels" {
			t.Errorf("expected /namespaces/test/resources/myres/channels, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.channel.v0+json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"namespace":"test","resource":"myres","name":"stable","version":{"namespace":"test","resource":"myres","string":"1.0.0","description":"","hasArchive":true,"createdAt":0,"updatedAt":0},"description":"","createdAt":0,"updatedAt":0}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	ch, err := client.CreateChannel(context.Background(), "test", "myres", ChannelInfo{
		Name:        "stable",
		Version:     "1.0.0",
		Description: "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.Name != "stable" {
		t.Errorf("expected channel 'stable', got %s", ch.Name)
	}
	if ch.Version.String != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", ch.Version.String)
	}
}

func TestClient_ReadChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres/channels/stable" {
			t.Errorf("expected /namespaces/test/resources/myres/channels/stable, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.channel.v0+json")
		w.Write([]byte(`{"namespace":"test","resource":"myres","name":"stable","version":{"namespace":"test","resource":"myres","string":"1.0.0","description":"","hasArchive":true,"createdAt":0,"updatedAt":0},"description":"","createdAt":0,"updatedAt":0}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	ch, err := client.ReadChannel(context.Background(), "test", "myres", "stable")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.Name != "stable" {
		t.Errorf("expected channel 'stable', got %s", ch.Name)
	}
}

func TestClient_UpdateChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres/channels/stable" {
			t.Errorf("expected /namespaces/test/resources/myres/channels/stable, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.channel.v0+json")
		w.Write([]byte(`{"namespace":"test","resource":"myres","name":"stable","version":{"namespace":"test","resource":"myres","string":"1.0.1","description":"","hasArchive":true,"createdAt":0,"updatedAt":0},"description":"updated","createdAt":0,"updatedAt":1}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	ch, err := client.UpdateChannel(context.Background(), "test", "myres", "stable", ChannelInfo{
		Name:        "stable",
		Version:     "1.0.1",
		Description: "updated",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.Description != "updated" {
		t.Errorf("expected description 'updated', got %s", ch.Description)
	}
	if ch.Version.String != "1.0.1" {
		t.Errorf("expected version '1.0.1', got %s", ch.Version.String)
	}
}

func TestClient_DeleteChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres/channels/stable" {
			t.Errorf("expected /namespaces/test/resources/myres/channels/stable, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	err := client.DeleteChannel(context.Background(), "test", "myres", "stable")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ListChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/namespaces/test/resources/myres/channels" {
			t.Errorf("expected /namespaces/test/resources/myres/channels, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.crucible.channel-list.v0+json")
		w.Write([]byte(`{"channels":[{"namespace":"test","resource":"myres","name":"stable","version":"1.0.0","description":"","createdAt":0,"updatedAt":0}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	list, err := client.ListChannels(context.Background(), "test", "myres")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Channels) != 1 {
		t.Errorf("expected 1 channel, got %d", len(list.Channels))
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.crucible.error.v0+json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"code":"not_found","message":"namespace not found"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	_, err := client.ReadNamespace(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNotFound {
		t.Errorf("expected not_found, got %s", regErr.Code)
	}
}

func TestClient_InvalidBaseURL(t *testing.T) {
	client := NewClient("://invalid", nil)
	_, err := client.ReadNamespace(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestNewClient_DefaultHTTPClient(t *testing.T) {
	client := NewClient("https://example.com", nil)
	if client.httpClient != http.DefaultClient {
		t.Error("expected default HTTP client when nil is passed")
	}
}

func TestNewClient_CustomHTTPClient(t *testing.T) {
	custom := &http.Client{}
	client := NewClient("https://example.com", custom)
	if client.httpClient != custom {
		t.Error("expected custom HTTP client to be used")
	}
}
