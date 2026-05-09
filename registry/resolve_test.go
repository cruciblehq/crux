package registry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cruciblehq/crux/reference"
)

// mustParseVersionConstraint parses a version constraint string and fatally fails the test on error.
func mustParseVersionConstraint(t *testing.T, s string) *reference.VersionConstraint {
	t.Helper()
	vc, err := reference.ParseVersionConstraint(s)
	if err != nil {
		t.Fatalf("ParseVersionConstraint(%q) failed: %v", s, err)
	}
	return vc
}

// mustParseReference parses a reference string and fatally fails the test on error.
func mustParseReference(t *testing.T, s, typ string) *reference.Reference {
	t.Helper()
	ref, err := reference.Parse(s, typ)
	if err != nil {
		t.Fatalf("reference.Parse(%q, %q) failed: %v", s, typ, err)
	}
	return ref
}

func TestFindLatestVersion_Empty(t *testing.T) {
	vc := mustParseVersionConstraint(t, "^1.0.0")
	result := FindLatestVersion(nil, vc)
	if result != nil {
		t.Error("expected nil for empty version list")
	}
}

func TestFindLatestVersion_NoMatch(t *testing.T) {
	versions := []VersionSummary{
		{String: "2.0.0", CreatedAt: 1, UpdatedAt: 1},
	}
	vc := mustParseVersionConstraint(t, "^1.0.0")
	result := FindLatestVersion(versions, vc)
	if result != nil {
		t.Error("expected nil when no versions match constraint")
	}
}

func TestFindLatestVersion_SingleMatch(t *testing.T) {
	versions := []VersionSummary{
		{String: "1.2.0", CreatedAt: 1, UpdatedAt: 1},
	}
	vc := mustParseVersionConstraint(t, "^1.0.0")
	result := FindLatestVersion(versions, vc)
	if result == nil {
		t.Fatal("expected a match, got nil")
	}
	if result.String() != "1.2.0" {
		t.Errorf("expected 1.2.0, got %s", result.String())
	}
}

func TestFindLatestVersion_PicksHighest(t *testing.T) {
	versions := []VersionSummary{
		{String: "1.0.0", CreatedAt: 1, UpdatedAt: 1},
		{String: "1.3.0", CreatedAt: 1, UpdatedAt: 1},
		{String: "1.1.0", CreatedAt: 1, UpdatedAt: 1},
	}
	vc := mustParseVersionConstraint(t, "^1.0.0")
	result := FindLatestVersion(versions, vc)
	if result == nil {
		t.Fatal("expected a match, got nil")
	}
	if result.String() != "1.3.0" {
		t.Errorf("expected 1.3.0, got %s", result.String())
	}
}

func TestFindLatestVersion_SkipsMalformedVersions(t *testing.T) {
	versions := []VersionSummary{
		{String: "not-a-version", CreatedAt: 1, UpdatedAt: 1},
		{String: "1.2.0", CreatedAt: 1, UpdatedAt: 1},
	}
	vc := mustParseVersionConstraint(t, "^1.0.0")
	result := FindLatestVersion(versions, vc)
	if result == nil {
		t.Fatal("expected a match, got nil")
	}
	if result.String() != "1.2.0" {
		t.Errorf("expected 1.2.0, got %s", result.String())
	}
}

func TestFindLatestVersion_ExcludesOutOfRangeVersions(t *testing.T) {
	versions := []VersionSummary{
		{String: "1.0.0", CreatedAt: 1, UpdatedAt: 1},
		{String: "2.0.0", CreatedAt: 1, UpdatedAt: 1},
		{String: "1.5.0", CreatedAt: 1, UpdatedAt: 1},
	}
	vc := mustParseVersionConstraint(t, "^1.0.0")
	result := FindLatestVersion(versions, vc)
	if result == nil {
		t.Fatal("expected a match, got nil")
	}
	if result.String() != "1.5.0" {
		t.Errorf("expected 1.5.0, got %s", result.String())
	}
}

func TestFindLatestVersion_ExactConstraint(t *testing.T) {
	versions := []VersionSummary{
		{String: "1.0.0", CreatedAt: 1, UpdatedAt: 1},
		{String: "1.1.0", CreatedAt: 1, UpdatedAt: 1},
	}
	vc := mustParseVersionConstraint(t, "1.0.0")
	result := FindLatestVersion(versions, vc)
	if result == nil {
		t.Fatal("expected a match, got nil")
	}
	if result.String() != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", result.String())
	}
}

// resolveServer builds a test HTTP server that routes requests by path and
// returns the provided JSON body with HTTP 200.
func resolveServer(t *testing.T, routes map[string]string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	for path, body := range routes {
		path, body := path, body
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, body)
		})
	}
	return httptest.NewServer(mux)
}

func TestResolveVersion_VersionConstraint_OK(t *testing.T) {
	server := resolveServer(t, map[string]string{
		"/namespaces/ns/resources/myresource": `{
			"namespace":"ns","name":"myresource","type":"widget",
			"versions":[],"channels":[],"createdAt":1,"updatedAt":1
		}`,
		"/namespaces/ns/resources/myresource/versions": `{
			"versions":[
				{"string":"1.0.0","createdAt":1,"updatedAt":1},
				{"string":"1.3.0","createdAt":1,"updatedAt":1}
			]
		}`,
		"/namespaces/ns/resources/myresource/versions/1.3.0": `{
			"namespace":"ns","resource":"myresource","string":"1.3.0",
			"createdAt":1,"updatedAt":1
		}`,
	})
	defer server.Close()

	client := NewClient(server.URL, nil)
	ref := mustParseReference(t, "widget ns/myresource ^1.0.0", "widget")
	v, err := ResolveVersion(context.Background(), client, ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.String != "1.3.0" {
		t.Errorf("expected version 1.3.0, got %s", v.String)
	}
}

func TestResolveVersion_Channel_OK(t *testing.T) {
	server := resolveServer(t, map[string]string{
		"/namespaces/ns/resources/myresource": `{
			"namespace":"ns","name":"myresource","type":"widget",
			"versions":[],"channels":[],"createdAt":1,"updatedAt":1
		}`,
		"/namespaces/ns/resources/myresource/channels/stable": `{
			"namespace":"ns","resource":"myresource","name":"stable",
			"version":{
				"namespace":"ns","resource":"myresource","string":"1.2.0",
				"createdAt":1,"updatedAt":1
			},
			"createdAt":1,"updatedAt":1
		}`,
	})
	defer server.Close()

	client := NewClient(server.URL, nil)
	ref := mustParseReference(t, "widget ns/myresource :stable", "widget")
	v, err := ResolveVersion(context.Background(), client, ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.String != "1.2.0" {
		t.Errorf("expected version 1.2.0, got %s", v.String)
	}
}

func TestResolveVersion_TypeMismatch(t *testing.T) {
	server := resolveServer(t, map[string]string{
		"/namespaces/ns/resources/myresource": `{
			"namespace":"ns","name":"myresource","type":"service",
			"versions":[],"channels":[],"createdAt":1,"updatedAt":1
		}`,
	})
	defer server.Close()

	client := NewClient(server.URL, nil)
	ref := mustParseReference(t, "widget ns/myresource ^1.0.0", "widget")
	_, err := ResolveVersion(context.Background(), client, ref)
	if !errors.Is(err, ErrTypeMismatch) {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

func TestResolveVersion_NoVersions(t *testing.T) {
	server := resolveServer(t, map[string]string{
		"/namespaces/ns/resources/myresource": `{
			"namespace":"ns","name":"myresource","type":"widget",
			"versions":[],"channels":[],"createdAt":1,"updatedAt":1
		}`,
		"/namespaces/ns/resources/myresource/versions": `{"versions":[]}`,
	})
	defer server.Close()

	client := NewClient(server.URL, nil)
	ref := mustParseReference(t, "widget ns/myresource ^1.0.0", "widget")
	_, err := ResolveVersion(context.Background(), client, ref)
	if !errors.Is(err, ErrNoVersions) {
		t.Errorf("expected ErrNoVersions, got %v", err)
	}
}

func TestResolveVersion_NoMatchingVersion(t *testing.T) {
	server := resolveServer(t, map[string]string{
		"/namespaces/ns/resources/myresource": `{
			"namespace":"ns","name":"myresource","type":"widget",
			"versions":[],"channels":[],"createdAt":1,"updatedAt":1
		}`,
		"/namespaces/ns/resources/myresource/versions": `{
			"versions":[{"string":"2.0.0","createdAt":1,"updatedAt":1}]
		}`,
	})
	defer server.Close()

	client := NewClient(server.URL, nil)
	ref := mustParseReference(t, "widget ns/myresource ^1.0.0", "widget")
	_, err := ResolveVersion(context.Background(), client, ref)
	if !errors.Is(err, ErrNoMatchingVersion) {
		t.Errorf("expected ErrNoMatchingVersion, got %v", err)
	}
}

func TestResolveVersion_ReadResourceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	ref := mustParseReference(t, "widget ns/myresource ^1.0.0", "widget")
	_, err := ResolveVersion(context.Background(), client, ref)
	if !errors.Is(err, ErrHTTPStatus) {
		t.Errorf("expected ErrHTTPStatus, got %v", err)
	}
}
