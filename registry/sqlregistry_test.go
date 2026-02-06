package registry

import (
	"bytes"
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewSQLRegistry(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "registry_test_*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	db, err := sql.Open("sqlite3", tmpfile.Name()+"?_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	ctx := context.Background()
	registry, err := NewSQLRegistry(ctx, db, tmpDir, logger)

	if err != nil {
		t.Fatalf("NewSQLRegistry() error = %v", err)
	}

	if registry == nil {
		t.Fatal("NewSQLRegistry() returned nil registry")
	}

	if registry.db != db {
		t.Error("registry.db should match provided db")
	}

	if registry.logger != logger {
		t.Error("registry.logger should match provided logger")
	}

	if registry.archiveRoot != tmpDir {
		t.Errorf("registry.archiveRoot = %q, want %q", registry.archiveRoot, tmpDir)
	}
}

func TestNewSQLRegistry_SchemaCreationError(t *testing.T) {
	// Create a database that will fail schema creation
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Close the database to force an error
	db.Close()

	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ctx := context.Background()

	_, err = NewSQLRegistry(ctx, db, tmpDir, logger)

	if err == nil {
		t.Error("NewSQLRegistry() expected error for closed database, got nil")
	}

	if regErr, ok := err.(*Error); ok {
		if regErr.Code != ErrorCodeInternalError {
			t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeInternalError)
		}
	} else {
		t.Errorf("expected *Error type, got %T", err)
	}
}

func TestCreateNamespace_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	info := NamespaceInfo{
		Name:        "test-namespace",
		Description: "Test namespace",
	}

	ns, err := registry.CreateNamespace(ctx, info)
	if err != nil {
		t.Fatalf("CreateNamespace() error = %v", err)
	}

	if ns.Name != info.Name {
		t.Errorf("Name = %q, want %q", ns.Name, info.Name)
	}
	if len(ns.Resources) != 0 {
		t.Error("new namespace should have empty resources list")
	}
}

func TestCreateNamespace_InvalidName(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	info := NamespaceInfo{
		Name:        "Invalid-Name",
		Description: "Test",
	}

	_, err := registry.CreateNamespace(ctx, info)
	if err == nil {
		t.Fatal("expected error for invalid name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestCreateNamespace_Duplicate(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	info := NamespaceInfo{Name: "test-ns", Description: "Test"}

	_, err := registry.CreateNamespace(ctx, info)
	if err != nil {
		t.Fatalf("first create error = %v", err)
	}

	_, err = registry.CreateNamespace(ctx, info)
	if err == nil {
		t.Fatal("expected error for duplicate namespace, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNamespaceExists {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeNamespaceExists)
	}
}

func TestReadNamespace_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})

	ns, err := registry.ReadNamespace(ctx, "test-ns")
	if err != nil {
		t.Fatalf("ReadNamespace() error = %v", err)
	}

	if ns.Name != "test-ns" {
		t.Errorf("Name = %q, want 'test-ns'", ns.Name)
	}
}

func TestReadNamespace_InvalidName(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.ReadNamespace(ctx, "Invalid-Name")
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestReadNamespace_NotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.ReadNamespace(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent namespace, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNotFound {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeNotFound)
	}
}

func TestUpdateNamespace_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Original"})

	updated, err := registry.UpdateNamespace(ctx, "test-ns", NamespaceInfo{Name: "test-ns", Description: "Updated"})
	if err != nil {
		t.Fatalf("UpdateNamespace() error = %v", err)
	}

	if updated.Description != "Updated" {
		t.Errorf("Description = %q, want 'Updated'", updated.Description)
	}
}

func TestUpdateNamespace_InvalidName(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.UpdateNamespace(ctx, "Invalid-Name", NamespaceInfo{Name: "test", Description: "Test"})
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestUpdateNamespace_NotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.UpdateNamespace(ctx, "nonexistent", NamespaceInfo{Name: "nonexistent", Description: "Test"})
	if err == nil {
		t.Fatal("expected error for nonexistent namespace, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNotFound {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeNotFound)
	}
}

func TestDeleteNamespace_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})

	err := registry.DeleteNamespace(ctx, "test-ns")
	if err != nil {
		t.Fatalf("DeleteNamespace() error = %v", err)
	}

	_, err = registry.ReadNamespace(ctx, "test-ns")
	if err == nil {
		t.Error("namespace should be deleted")
	}
}

func TestDeleteNamespace_InvalidName(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := registry.DeleteNamespace(ctx, "Invalid-Name")
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestDeleteNamespace_WithResources(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})

	err := registry.DeleteNamespace(ctx, "test-ns")
	if err == nil {
		t.Fatal("expected error deleting namespace with resources, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	// Note: Current implementation returns InternalError for FK constraint violations
	// rather than NamespaceNotEmpty. This could be improved in the future.
	if regErr.Code != ErrorCodeInternalError {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeInternalError)
	}
}

func TestListNamespaces(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "ns1", Description: "First"})
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "ns2", Description: "Second"})

	list, err := registry.ListNamespaces(ctx)
	if err != nil {
		t.Fatalf("ListNamespaces() error = %v", err)
	}

	if len(list.Namespaces) != 2 {
		t.Errorf("expected 2 namespaces, got %d", len(list.Namespaces))
	}
}

func TestCreateResource_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})

	info := ResourceInfo{
		Name:        "test-resource",
		Type:        "widget",
		Description: "Test resource",
	}

	res, err := registry.CreateResource(ctx, "test-ns", info)
	if err != nil {
		t.Fatalf("CreateResource() error = %v", err)
	}

	if res.Name != info.Name {
		t.Errorf("Name = %q, want %q", res.Name, info.Name)
	}
}

func TestCreateResource_NamespaceNotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	info := ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"}

	_, err := registry.CreateResource(ctx, "nonexistent", info)
	if err == nil {
		t.Fatal("expected error creating resource in nonexistent namespace, got nil")
	}
}

func TestCreateResource_InvalidNamespace(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	info := ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"}

	_, err := registry.CreateResource(ctx, "Invalid-Name", info)
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestCreateResource_InvalidResource(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	info := ResourceInfo{Name: "Invalid-Name", Type: "widget", Description: "Test"}

	_, err := registry.CreateResource(ctx, "test-ns", info)
	if err == nil {
		t.Fatal("expected error for invalid resource name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestCreateResource_Duplicate(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	info := ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"}

	_, err := registry.CreateResource(ctx, "test-ns", info)
	if err != nil {
		t.Fatalf("first create error = %v", err)
	}

	_, err = registry.CreateResource(ctx, "test-ns", info)
	if err == nil {
		t.Fatal("expected error for duplicate resource, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeResourceExists {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeResourceExists)
	}
}

func TestReadResource_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})

	res, err := registry.ReadResource(ctx, "test-ns", "test-resource")
	if err != nil {
		t.Fatalf("ReadResource() error = %v", err)
	}

	if res.Name != "test-resource" {
		t.Errorf("Name = %q, want 'test-resource'", res.Name)
	}
}

func TestReadResource_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.ReadResource(ctx, "Invalid-Name", "test-resource")
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestReadResource_NotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})

	_, err := registry.ReadResource(ctx, "test-ns", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent resource, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNotFound {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeNotFound)
	}
}

func TestUpdateResource_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Original"})

	updated, err := registry.UpdateResource(ctx, "test-ns", "test-resource", ResourceInfo{Name: "test-resource", Type: "service", Description: "Updated"})
	if err != nil {
		t.Fatalf("UpdateResource() error = %v", err)
	}

	if updated.Type != "service" {
		t.Errorf("Type = %q, want 'service'", updated.Type)
	}
}

func TestUpdateResource_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.UpdateResource(ctx, "Invalid-Name", "test-resource", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestUpdateResource_NotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})

	_, err := registry.UpdateResource(ctx, "test-ns", "nonexistent", ResourceInfo{Name: "nonexistent", Type: "widget", Description: "Test"})
	if err == nil {
		t.Fatal("expected error for nonexistent resource, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNotFound {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeNotFound)
	}
}

func TestDeleteResource_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})

	err := registry.DeleteResource(ctx, "test-ns", "test-resource")
	if err != nil {
		t.Fatalf("DeleteResource() error = %v", err)
	}
}

func TestDeleteResource_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := registry.DeleteResource(ctx, "Invalid-Name", "test-resource")
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestListResources(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "res1", Type: "widget", Description: "First"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "res2", Type: "service", Description: "Second"})

	list, err := registry.ListResources(ctx, "test-ns")
	if err != nil {
		t.Fatalf("ListResources() error = %v", err)
	}

	if len(list.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(list.Resources))
	}
}

func TestListResources_InvalidName(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.ListResources(ctx, "Invalid-Name")
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestCreateVersion_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})

	info := VersionInfo{String: "1.0.0"}
	v, err := registry.CreateVersion(ctx, "test-ns", "test-resource", info)
	if err != nil {
		t.Fatalf("CreateVersion() error = %v", err)
	}

	if v.String != "1.0.0" {
		t.Errorf("String = %q, want '1.0.0'", v.String)
	}
}

func TestCreateVersion_InvalidVersion(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})

	info := VersionInfo{String: "not-a-version"}
	_, err := registry.CreateVersion(ctx, "test-ns", "test-resource", info)
	if err == nil {
		t.Fatal("expected error for invalid version, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestCreateVersion_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	info := VersionInfo{String: "1.0.0"}
	_, err := registry.CreateVersion(ctx, "Invalid-Name", "test-resource", info)
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestCreateVersion_ResourceNotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})

	info := VersionInfo{String: "1.0.0"}
	_, err := registry.CreateVersion(ctx, "test-ns", "nonexistent", info)
	if err == nil {
		t.Fatal("expected error for nonexistent resource, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNotFound {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeNotFound)
	}
}

func TestCreateVersion_Duplicate(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})

	info := VersionInfo{String: "1.0.0"}
	_, err := registry.CreateVersion(ctx, "test-ns", "test-resource", info)
	if err != nil {
		t.Fatalf("first create error = %v", err)
	}

	_, err = registry.CreateVersion(ctx, "test-ns", "test-resource", info)
	if err == nil {
		t.Fatal("expected error for duplicate version, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeVersionExists {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeVersionExists)
	}
}

func TestReadVersion_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})

	v, err := registry.ReadVersion(ctx, "test-ns", "test-resource", "1.0.0")
	if err != nil {
		t.Fatalf("ReadVersion() error = %v", err)
	}

	if v.String != "1.0.0" {
		t.Errorf("String = %q, want '1.0.0'", v.String)
	}
}

func TestReadVersion_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.ReadVersion(ctx, "Invalid-Name", "test-resource", "1.0.0")
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestReadVersion_NotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})

	_, err := registry.ReadVersion(ctx, "test-ns", "test-resource", "9.9.9")
	if err == nil {
		t.Fatal("expected error for nonexistent version, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNotFound {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeNotFound)
	}
}

func TestUpdateVersion_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})

	updated, err := registry.UpdateVersion(ctx, "test-ns", "test-resource", "1.0.0", VersionInfo{String: "1.0.0"})
	if err != nil {
		t.Fatalf("UpdateVersion() error = %v", err)
	}

	if updated.String != "1.0.0" {
		t.Errorf("String = %q, want '1.0.0'", updated.String)
	}
}

func TestUpdateVersion_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.UpdateVersion(ctx, "Invalid-Name", "test-resource", "1.0.0", VersionInfo{String: "1.0.0"})
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestUpdateVersion_NotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})

	_, err := registry.UpdateVersion(ctx, "test-ns", "test-resource", "9.9.9", VersionInfo{String: "9.9.9"})
	if err == nil {
		t.Fatal("expected error for nonexistent version, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNotFound {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeNotFound)
	}
}

func TestDeleteVersion_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})

	err := registry.DeleteVersion(ctx, "test-ns", "test-resource", "1.0.0")
	if err != nil {
		t.Fatalf("DeleteVersion() error = %v", err)
	}

	_, err = registry.ReadVersion(ctx, "test-ns", "test-resource", "1.0.0")
	if err == nil {
		t.Error("version should be deleted")
	}
}

func TestDeleteVersion_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := registry.DeleteVersion(ctx, "Invalid-Name", "test-resource", "1.0.0")
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestListVersions(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.1.0"})

	list, err := registry.ListVersions(ctx, "test-ns", "test-resource")
	if err != nil {
		t.Fatalf("ListVersions() error = %v", err)
	}

	if len(list.Versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(list.Versions))
	}
}

func TestListVersions_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.ListVersions(ctx, "Invalid-Name", "test-resource")
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestUploadArchive_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})

	archiveData := []byte("test archive content")
	reader := bytes.NewReader(archiveData)

	v, err := registry.UploadArchive(ctx, "test-ns", "test-resource", "1.0.0", reader)
	if err != nil {
		t.Fatalf("UploadArchive() error = %v", err)
	}

	if v.Digest == nil {
		t.Error("Digest should be set after upload")
	}
	if v.Size == nil {
		t.Error("Size should be set after upload")
	}
	if v.Archive == nil {
		t.Error("Archive path should be set after upload")
	}
	if *v.Size != int64(len(archiveData)) {
		t.Errorf("Size = %d, want %d", *v.Size, len(archiveData))
	}
}

func TestUploadArchive_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	archiveData := []byte("test")
	reader := bytes.NewReader(archiveData)

	_, err := registry.UploadArchive(ctx, "Invalid-Name", "test-resource", "1.0.0", reader)
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestDownloadArchive_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})

	archiveData := []byte("test archive content")
	_, _ = registry.UploadArchive(ctx, "test-ns", "test-resource", "1.0.0", bytes.NewReader(archiveData))

	reader, err := registry.DownloadArchive(ctx, "test-ns", "test-resource", "1.0.0")
	if err != nil {
		t.Fatalf("DownloadArchive() error = %v", err)
	}
	defer reader.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader)
	if err != nil {
		t.Fatalf("failed to read archive: %v", err)
	}

	if buf.String() != string(archiveData) {
		t.Errorf("downloaded content doesn't match uploaded content")
	}
}

func TestDownloadArchive_NotUploaded(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})

	_, err := registry.DownloadArchive(ctx, "test-ns", "test-resource", "1.0.0")
	if err == nil {
		t.Fatal("expected error downloading archive that hasn't been uploaded, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNotFound {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeNotFound)
	}
}

func TestDownloadArchive_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.DownloadArchive(ctx, "Invalid-Name", "test-resource", "1.0.0")
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestDownloadArchive_VersionNotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})

	_, err := registry.DownloadArchive(ctx, "test-ns", "test-resource", "9.9.9")
	if err == nil {
		t.Fatal("expected error for nonexistent version, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNotFound {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeNotFound)
	}
}

func TestCreateChannel_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})

	info := ChannelInfo{
		Name:        "stable",
		Version:     "1.0.0",
		Description: "Stable channel",
	}

	ch, err := registry.CreateChannel(ctx, "test-ns", "test-resource", info)
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}

	if ch.Name != "stable" {
		t.Errorf("Name = %q, want 'stable'", ch.Name)
	}
	if ch.Version.String != "1.0.0" {
		t.Errorf("Version.String = %q, want '1.0.0'", ch.Version.String)
	}
}

func TestCreateChannel_InvalidName(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})

	info := ChannelInfo{Name: "Invalid-Name", Version: "1.0.0", Description: "Test"}

	_, err := registry.CreateChannel(ctx, "test-ns", "test-resource", info)
	if err == nil {
		t.Fatal("expected error for invalid channel name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestCreateChannel_InvalidNamespace(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	info := ChannelInfo{Name: "stable", Version: "1.0.0", Description: "Test"}
	_, err := registry.CreateChannel(ctx, "Invalid-Name", "test-resource", info)
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestCreateChannel_Duplicate(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})

	info := ChannelInfo{Name: "stable", Version: "1.0.0", Description: "Test"}
	_, err := registry.CreateChannel(ctx, "test-ns", "test-resource", info)
	if err != nil {
		t.Fatalf("first create error = %v", err)
	}

	_, err = registry.CreateChannel(ctx, "test-ns", "test-resource", info)
	if err == nil {
		t.Fatal("expected error for duplicate channel, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeChannelExists {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeChannelExists)
	}
}

func TestCreateChannel_ResourceNotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})

	info := ChannelInfo{Name: "stable", Version: "1.0.0", Description: "Test"}
	_, err := registry.CreateChannel(ctx, "test-ns", "nonexistent-resource", info)
	if err == nil {
		t.Fatal("expected error for nonexistent resource, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNotFound {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeNotFound)
	}
}

func TestUpdateChannel_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "2.0.0"})
	_, _ = registry.CreateChannel(ctx, "test-ns", "test-resource", ChannelInfo{Name: "stable", Version: "1.0.0", Description: "Stable"})

	updated, err := registry.UpdateChannel(ctx, "test-ns", "test-resource", "stable", ChannelInfo{Name: "stable", Version: "2.0.0", Description: "Updated"})
	if err != nil {
		t.Fatalf("UpdateChannel() error = %v", err)
	}

	if updated.Version.String != "2.0.0" {
		t.Errorf("Version.String = %q, want '2.0.0'", updated.Version.String)
	}
}

func TestUpdateChannel_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.UpdateChannel(ctx, "Invalid-Name", "test-resource", "stable", ChannelInfo{Name: "stable", Version: "1.0.0", Description: "Test"})
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestUpdateChannel_NotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})

	_, err := registry.UpdateChannel(ctx, "test-ns", "test-resource", "nonexistent", ChannelInfo{Name: "nonexistent", Version: "1.0.0", Description: "Test"})
	if err == nil {
		t.Fatal("expected error for nonexistent channel, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNotFound {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeNotFound)
	}
}

func TestReadChannel_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})
	_, _ = registry.CreateChannel(ctx, "test-ns", "test-resource", ChannelInfo{Name: "stable", Version: "1.0.0", Description: "Stable"})

	ch, err := registry.ReadChannel(ctx, "test-ns", "test-resource", "stable")
	if err != nil {
		t.Fatalf("ReadChannel() error = %v", err)
	}

	if ch.Name != "stable" {
		t.Errorf("Name = %q, want 'stable'", ch.Name)
	}
}

func TestReadChannel_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.ReadChannel(ctx, "Invalid-Name", "test-resource", "stable")
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestReadChannel_NotFound(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})

	_, err := registry.ReadChannel(ctx, "test-ns", "test-resource", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent channel, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeNotFound {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeNotFound)
	}
}

func TestDeleteChannel_Success(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})
	_, _ = registry.CreateChannel(ctx, "test-ns", "test-resource", ChannelInfo{Name: "stable", Version: "1.0.0", Description: "Stable"})

	err := registry.DeleteChannel(ctx, "test-ns", "test-resource", "stable")
	if err != nil {
		t.Fatalf("DeleteChannel() error = %v", err)
	}
}

func TestDeleteChannel_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := registry.DeleteChannel(ctx, "Invalid-Name", "test-resource", "stable")
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}

func TestListChannels(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = registry.CreateNamespace(ctx, NamespaceInfo{Name: "test-ns", Description: "Test"})
	_, _ = registry.CreateResource(ctx, "test-ns", ResourceInfo{Name: "test-resource", Type: "widget", Description: "Test"})
	_, _ = registry.CreateVersion(ctx, "test-ns", "test-resource", VersionInfo{String: "1.0.0"})
	_, _ = registry.CreateChannel(ctx, "test-ns", "test-resource", ChannelInfo{Name: "stable", Version: "1.0.0", Description: "Stable"})
	_, _ = registry.CreateChannel(ctx, "test-ns", "test-resource", ChannelInfo{Name: "beta", Version: "1.0.0", Description: "Beta"})

	list, err := registry.ListChannels(ctx, "test-ns", "test-resource")
	if err != nil {
		t.Fatalf("ListChannels() error = %v", err)
	}

	if len(list.Channels) != 2 {
		t.Errorf("expected 2 channels, got %d", len(list.Channels))
	}
}

func TestListChannels_InvalidNames(t *testing.T) {
	registry, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := registry.ListChannels(ctx, "Invalid-Name", "test-resource")
	if err == nil {
		t.Fatal("expected error for invalid namespace name, got nil")
	}

	regErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if regErr.Code != ErrorCodeBadRequest {
		t.Errorf("error code = %v, want %v", regErr.Code, ErrorCodeBadRequest)
	}
}
