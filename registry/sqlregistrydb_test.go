package registry

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Creates a temporary database for testing
func setupTestDB(t *testing.T) (*SQLRegistry, func()) {
	t.Helper()

	// Create temp database file
	tmpfile, err := os.CreateTemp("", "registry_test_*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Open database
	db, err := sql.Open("sqlite3", tmpfile.Name()+"?_foreign_keys=on")
	if err != nil {
		os.Remove(tmpfile.Name())
		t.Fatal(err)
	}

	// Create schema
	_, err = db.Exec(sqlSchema)
	if err != nil {
		db.Close()
		os.Remove(tmpfile.Name())
		t.Fatalf("failed to create schema: %v", err)
	}

	// Create registry
	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	registry := &SQLRegistry{
		db:          db,
		logger:      logger,
		archiveRoot: tmpDir,
	}

	// Cleanup function
	cleanup := func() {
		db.Close()
		os.Remove(tmpfile.Name())
	}

	return registry, cleanup
}

func TestInsertDBNamespace(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	info := NamespaceInfo{
		Name:        "test-ns",
		Description: "Test namespace",
	}

	ns, err := registry.insertNamespace(ctx, info)
	if err != nil {
		t.Fatalf("insertNamespace() error = %v", err)
	}

	if ns.Name != info.Name {
		t.Errorf("Name = %q, want %q", ns.Name, info.Name)
	}
	if ns.Description != info.Description {
		t.Errorf("Description = %q, want %q", ns.Description, info.Description)
	}
	if ns.CreatedAt == 0 {
		t.Error("CreatedAt should be set")
	}
	if ns.UpdatedAt == 0 {
		t.Error("UpdatedAt should be set")
	}
	if len(ns.Resources) != 0 {
		t.Errorf("Resources should be empty, got %d items", len(ns.Resources))
	}
}

func TestInsertDBNamespace_Duplicate(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	info := NamespaceInfo{Name: "test-ns", Description: "Test"}

	_, err := registry.insertNamespace(ctx, info)
	if err != nil {
		t.Fatalf("first insert error = %v", err)
	}

	_, err = registry.insertNamespace(ctx, info)
	if err == nil {
		t.Error("expected error for duplicate namespace, got nil")
	}
}

func TestGetDBNamespace(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Insert namespace
	info := NamespaceInfo{Name: "test-ns", Description: "Test"}
	inserted, err := registry.insertNamespace(ctx, info)
	if err != nil {
		t.Fatalf("insertNamespace() error = %v", err)
	}

	// Get namespace
	ns, err := registry.getNamespace(ctx, "test-ns")
	if err != nil {
		t.Fatalf("getNamespace() error = %v", err)
	}

	if ns.Name != inserted.Name {
		t.Errorf("Name = %q, want %q", ns.Name, inserted.Name)
	}
	if ns.Description != inserted.Description {
		t.Errorf("Description = %q, want %q", ns.Description, inserted.Description)
	}
}

func TestGetDBNamespace_NotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.getNamespace(ctx, "nonexistent")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestUpdateDBNamespace(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Insert namespace
	insertInfo := NamespaceInfo{Name: "test-ns", Description: "Original"}
	inserted, err := registry.insertNamespace(ctx, insertInfo)
	if err != nil {
		t.Fatalf("insertNamespace() error = %v", err)
	}

	// Wait to ensure updated_at changes
	time.Sleep(100 * time.Millisecond)

	// Update namespace
	updateInfo := NamespaceInfo{Name: "test-ns", Description: "Updated"}
	updated, err := registry.updateNamespace(ctx, "test-ns", updateInfo)
	if err != nil {
		t.Fatalf("updateNamespace() error = %v", err)
	}

	if updated.Description != "Updated" {
		t.Errorf("Description = %q, want 'Updated'", updated.Description)
	}
	if updated.CreatedAt != inserted.CreatedAt {
		t.Error("CreatedAt should not change")
	}
}

func TestUpdateDBNamespace_NotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	info := NamespaceInfo{Name: "test", Description: "Test"}

	_, err := registry.updateNamespace(ctx, "nonexistent", info)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestDeleteDBNamespace(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Insert namespace
	info := NamespaceInfo{Name: "test-ns", Description: "Test"}
	_, err := registry.insertNamespace(ctx, info)
	if err != nil {
		t.Fatalf("insertNamespace() error = %v", err)
	}

	// Delete namespace
	err = registry.deleteNamespace(ctx, "test-ns")
	if err != nil {
		t.Fatalf("deleteNamespace() error = %v", err)
	}

	// Verify it's gone
	_, err = registry.getNamespace(ctx, "test-ns")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

func TestInsertDBResource(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace first
	nsInfo := NamespaceInfo{Name: "test-ns", Description: "Test"}
	_, err := registry.insertNamespace(ctx, nsInfo)
	if err != nil {
		t.Fatalf("insertNamespace() error = %v", err)
	}

	// Insert resource
	info := ResourceInfo{
		Name:        "test-resource",
		Type:        "widget",
		Description: "Test resource",
	}

	res, err := registry.insertResource(ctx, "test-ns", info)
	if err != nil {
		t.Fatalf("insertResource() error = %v", err)
	}

	if res.Name != info.Name {
		t.Errorf("Name = %q, want %q", res.Name, info.Name)
	}
	if res.Type != info.Type {
		t.Errorf("Type = %q, want %q", res.Type, info.Type)
	}
	if res.Description != info.Description {
		t.Errorf("Description = %q, want %q", res.Description, info.Description)
	}
	if res.Namespace != "test-ns" {
		t.Errorf("Namespace = %q, want 'test-ns'", res.Namespace)
	}
}

func TestGetDBResource(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace and resource
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	info := ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"}
	inserted, err := registry.insertResource(ctx, "test-ns", info)
	if err != nil {
		t.Fatalf("insertResource() error = %v", err)
	}

	// Get resource
	res, err := registry.getResource(ctx, "test-ns", "test-resource")
	if err != nil {
		t.Fatalf("getResource() error = %v", err)
	}

	if res.Name != inserted.Name {
		t.Errorf("Name = %q, want %q", res.Name, inserted.Name)
	}
	if res.Namespace != "test-ns" {
		t.Errorf("Namespace = %q, want 'test-ns'", res.Namespace)
	}
}

func TestUpdateDBResource(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace and resource
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	insertInfo := ResourceInfo{Name: "test-resource", Type: "widget", Description: "Original"}
	_, err := registry.insertResource(ctx, "test-ns", insertInfo)
	if err != nil {
		t.Fatalf("insertResource() error = %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	// Update resource
	updateInfo := ResourceInfo{Name: "test-resource", Type: "service", Description: "Updated"}
	updated, err := registry.updateResource(ctx, "test-ns", "test-resource", updateInfo)
	if err != nil {
		t.Fatalf("updateResource() error = %v", err)
	}

	if updated.Type != "service" {
		t.Errorf("Type = %q, want 'service'", updated.Type)
	}
	if updated.Description != "Updated" {
		t.Errorf("Description = %q, want 'Updated'", updated.Description)
	}
}

func TestDeleteDBResource(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace and resource
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	info := ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"}
	_, err := registry.insertResource(ctx, "test-ns", info)
	if err != nil {
		t.Fatalf("insertResource() error = %v", err)
	}

	// Delete resource
	err = registry.deleteResource(ctx, "test-ns", "test-resource")
	if err != nil {
		t.Fatalf("deleteResource() error = %v", err)
	}

	// Verify it's gone
	_, err = registry.getResource(ctx, "test-ns", "test-resource")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

func TestInsertDBVersion(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace and resource
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.insertResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})

	// Insert version
	info := VersionInfo{String: "1.0.0"}
	v, err := registry.insertVersion(ctx, "test-ns", "test-resource", info)
	if err != nil {
		t.Fatalf("insertVersion() error = %v", err)
	}

	if v.String != "1.0.0" {
		t.Errorf("String = %q, want '1.0.0'", v.String)
	}
	if v.Namespace != "test-ns" {
		t.Errorf("Namespace = %q, want 'test-ns'", v.Namespace)
	}
	if v.Resource != "test-resource" {
		t.Errorf("Resource = %q, want 'test-resource'", v.Resource)
	}
	if v.Digest != nil {
		t.Error("Digest should be nil initially")
	}
	if v.Size != nil {
		t.Error("Size should be nil initially")
	}
	if v.Archive != nil {
		t.Error("Archive should be nil initially")
	}
}

func TestGetDBVersion(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace, resource, and version
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.insertResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	info := VersionInfo{String: "1.0.0"}
	inserted, err := registry.insertVersion(ctx, "test-ns", "test-resource", info)
	if err != nil {
		t.Fatalf("insertVersion() error = %v", err)
	}

	// Get version
	v, err := registry.getVersion(ctx, "test-ns", "test-resource", "1.0.0")
	if err != nil {
		t.Fatalf("getVersion() error = %v", err)
	}

	if v.String != inserted.String {
		t.Errorf("String = %q, want %q", v.String, inserted.String)
	}
	if v.Namespace != "test-ns" {
		t.Errorf("Namespace = %q, want 'test-ns'", v.Namespace)
	}
}

func TestUpdateDBVersion(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace, resource, and version
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.insertResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, err := registry.insertVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})
	if err != nil {
		t.Fatalf("insertVersion() error = %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	// Update version
	_, err = registry.updateVersion(ctx, "test-ns", "test-resource", "1.0.0")
	if err != nil {
		t.Fatalf("updateVersion() error = %v", err)
	}
}

func TestDeleteDBVersion(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace, resource, and version
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.insertResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, err := registry.insertVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})
	if err != nil {
		t.Fatalf("insertVersion() error = %v", err)
	}

	// Delete version
	err = registry.deleteVersion(ctx, "test-ns", "test-resource", "1.0.0")
	if err != nil {
		t.Fatalf("deleteVersion() error = %v", err)
	}

	// Verify it's gone
	_, err = registry.getVersion(ctx, "test-ns", "test-resource", "1.0.0")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

func TestUploadDBArchive(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace, resource, and version
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.insertResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, err := registry.insertVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})
	if err != nil {
		t.Fatalf("insertVersion() error = %v", err)
	}

	// Upload archive metadata
	digest := "abc123"
	path := "/path/to/archive.tar.zst"
	size := int64(1024)

	err = registry.uploadArchive(ctx, "test-ns", "test-resource", "1.0.0", digest, path, size)
	if err != nil {
		t.Fatalf("uploadArchive() error = %v", err)
	}

	// Verify archive fields were set
	v, err := registry.getVersion(ctx, "test-ns", "test-resource", "1.0.0")
	if err != nil {
		t.Fatalf("getVersion() error = %v", err)
	}

	if v.Digest == nil || *v.Digest != digest {
		t.Errorf("Digest = %v, want %q", v.Digest, digest)
	}
	if v.Size == nil || *v.Size != size {
		t.Errorf("Size = %v, want %d", v.Size, size)
	}
	if v.Archive == nil || *v.Archive != path {
		t.Errorf("Archive = %v, want %q", v.Archive, path)
	}
}

func TestListDBNamespaces(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Insert multiple namespaces
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "ns1", Description: "First"})
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "ns2", Description: "Second"})

	// List namespaces
	namespaces, err := registry.listNamespaces(ctx)
	if err != nil {
		t.Fatalf("listNamespaces() error = %v", err)
	}

	if len(namespaces) != 2 {
		t.Errorf("expected 2 namespaces, got %d", len(namespaces))
	}
}

func TestListDBResources(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace and resources
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.insertResource(ctx, "test-ns", ResourceInfo{Name: "res1", Type: "widget", Description: "First"})
	_, _ = registry.insertResource(ctx, "test-ns", ResourceInfo{Name: "res2", Type: "service", Description: "Second"})

	// List resources
	resources, err := registry.listResources(ctx, "test-ns")
	if err != nil {
		t.Fatalf("listResources() error = %v", err)
	}

	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}
}

func TestListDBVersions(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace, resource, and versions
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.insertResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.insertVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})
	_, _ = registry.insertVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.1.0"})

	// List versions
	versions, err := registry.listVersions(ctx, "test-ns", "test-resource")
	if err != nil {
		t.Fatalf("listVersions() error = %v", err)
	}

	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}
}

func TestInsertDBChannel(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace, resource, and version
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.insertResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.insertVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})

	// Insert channel
	info := ChannelInfo{
		Name:        "stable",
		Version:     "1.0.0",
		Description: "Stable channel",
	}

	err := registry.insertChannel(ctx, "test-ns", "test-resource", info)
	if err != nil {
		t.Fatalf("insertChannel() error = %v", err)
	}

	// Verify channel exists
	ch, err := registry.getChannel(ctx, "test-ns", "test-resource", "stable")
	if err != nil {
		t.Fatalf("getChannel() error = %v", err)
	}

	if ch.Name != "stable" {
		t.Errorf("Name = %q, want 'stable'", ch.Name)
	}
	if ch.Version.String != "1.0.0" {
		t.Errorf("Version.String = %q, want '1.0.0'", ch.Version.String)
	}
}

func TestUpdateDBChannel(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace, resource, versions, and channel
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.insertResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.insertVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})
	_, _ = registry.insertVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "2.0.0"})
	_ = registry.insertChannel(ctx, "test-ns", "test-resource", ChannelInfo{Name: "stable", Version: "1.0.0", Description: "Stable"})

	time.Sleep(10 * time.Millisecond)

	// Update channel to point to new version
	updateInfo := ChannelInfo{Name: "stable", Version: "2.0.0", Description: "Updated stable"}
	updated, err := registry.updateChannel(ctx, "test-ns", "test-resource", updateInfo)
	if err != nil {
		t.Fatalf("updateChannel() error = %v", err)
	}

	if updated.Version.String != "2.0.0" {
		t.Errorf("Version.String = %q, want '2.0.0'", updated.Version.String)
	}
	if updated.Description != "Updated stable" {
		t.Errorf("Description = %q, want 'Updated stable'", updated.Description)
	}
}

func TestDeleteDBChannel(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace, resource, version, and channel
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.insertResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.insertVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})
	_ = registry.insertChannel(ctx, "test-ns", "test-resource", ChannelInfo{Name: "stable", Version: "1.0.0", Description: "Stable"})

	// Delete channel
	err := registry.deleteChannel(ctx, "test-ns", "test-resource", "stable")
	if err != nil {
		t.Fatalf("deleteChannel() error = %v", err)
	}

	// Verify it's gone
	_, err = registry.getChannel(ctx, "test-ns", "test-resource", "stable")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

func TestListDBChannels(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create namespace, resource, version, and channels
	_, _ = registry.insertNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.insertResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.insertVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})
	_ = registry.insertChannel(ctx, "test-ns", "test-resource", ChannelInfo{Name: "stable", Version: "1.0.0", Description: "Stable"})
	_ = registry.insertChannel(ctx, "test-ns", "test-resource", ChannelInfo{Name: "beta", Version: "1.0.0", Description: "Beta"})

	// List channels
	channels, err := registry.listChannels(ctx, "test-ns", "test-resource")
	if err != nil {
		t.Fatalf("listChannels() error = %v", err)
	}

	if len(channels) != 2 {
		t.Errorf("expected 2 channels, got %d", len(channels))
	}
}
